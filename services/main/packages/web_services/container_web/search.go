// Copyright (c) 2020 Wind River Systems, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at:
//       http://www.apache.org/licenses/LICENSE-2.0
// Unless required by applicable law or agreed to in writing, software  distributed
// under the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES
// OR CONDITIONS OF ANY KIND, either express or implied.

package container_web

import (
	"net/http"

	"strconv"

	"encoding/json"

	"time"

	"wrs/tk/packages/core/archive"
	"wrs/tk/packages/core/file_collection"

	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"gopkg.in/guregu/null.v4"
)

// ContainerSearchResult represents a potential match for a container query
type ContainerSearchResult struct {
	FileCollectionID         int64    `db:"id" json:"id"`
	Name                     string   `db:"name" json:"name"`
	Sha1                     string   `db:"checksum_sha1" json:"sha1"`
	FileCollectionInsertDate string   `db:"date" json:"date"`
	Count                    int64    `db:"count" json:"count"`
	Packages                 int64    `db:"packages" json:"packages"`
	Distance                 null.Int `db:"distance" json:"distance"`
}

// handleContainerSearch searches the database for containers matching the given queries.
// It can still handle a single post and response, but the expected handling is a websocket returning ContainerSearchResults one by one.
// See httpContainerSerach or webSocketContainerSearch.
//
// query is the string to search by.
// method is the search method, levenshtein is the default.
//
//	like, fast, and levenshtein are supported.
//
// depth is shallow or deep, shallow by default.
//
//	a shallow search only counts the number of files directly owned by the archive
//	a deep search also counts files in sub-archives
//
// auto is used by a websocket query
//
//	if auto, expect frontend to use the websocket to ask for deep file counts
func HandleContainerSearch(w http.ResponseWriter, r *http.Request) {
	queryValues := r.URL.Query()

	_, hasQuery := queryValues["query"]
	if !hasQuery {
		log.Trace().Msg("Missing url parameter: 'query'")
		http.Error(w, "Missing url parameter: 'query'", 400)
		return
	}
	searchQuery := queryValues.Get("query")

	var method archive.SearchMethod
	_, hasMethod := queryValues["method"]
	if !hasMethod {
		method = archive.METHOD_LEVENSHTEIN
	} else {
		method = archive.ParseMethod(queryValues.Get("method"))
	}

	depth := "shallow"
	if _, hasDepth := queryValues["depth"]; hasDepth {
		depth = queryValues.Get("depth")
	}

	upgradeToWebSocket := false
	for _, u := range r.Header.Values("Upgrade") {
		if u == "websocket" {
			upgradeToWebSocket = true
			break
		}
	}

	auto := false
	if _, hasAuto := queryValues["auto"]; hasAuto {
		auto, _ = strconv.ParseBool(queryValues.Get("auto"))
	}

	log.Debug().Str(zerolog.CallerFieldName, "handleContainerSearch").Str("query", searchQuery).Str("method", archive.SearchMethodString(method)).Str("depth", depth).Bool("websocket", upgradeToWebSocket).Bool("auto", auto).Send()

	archiveController, err := archive.GetArchiveController(r.Context())
	if err != nil {
		http.Error(w, "error getting archive controller", 500)
		return
	}

	if upgradeToWebSocket {
		if err := webSocketContainerSearch(w, r, archiveController, searchQuery, method, depth, auto); err != nil {
			log.Error().Err(err).Msg("error streaming archive search")
			return
		}

		return
	}

	results, err := archiveController.SearchForArchiveAll(searchQuery, method)
	if err != nil {
		http.Error(w, "error searching for archive", 500)
		return
	}

	if err := json.NewEncoder(w).Encode(results); err != nil {
		http.Error(w, "error encoding JSON response", 500)
		return
	}
}

// webSocketContainerSearch searches for containers using the given query, and method, including a deep or shallow file count, and returns the results one by one over a websocket.
// After it is done it returns the total row count.
// If auto is specified, it then responds to further requests over the websocket, and responds with deep file counts.
func webSocketContainerSearch(w http.ResponseWriter, r *http.Request, archiveController *archive.ArchiveController, query string, method archive.SearchMethod, depth string, auto bool) error {
	// Handle connection
	log.Trace().Str(zerolog.CallerFieldName, "webSocketContainerSearch").Str("query", query).Str("method", archive.SearchMethodString(method)).Str("depth", depth).Bool("auto", auto).Send()
	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024}
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return errors.Wrapf(err, "unable to upgrade to websocket connection")
	}
	defer ws.Close()

	go func() { // Send pings to receive pongs that keep connection alive
		err := ws.WriteControl(websocket.PingMessage, nil, time.Now().Add(time.Minute))
		for err == nil {
			time.Sleep(time.Second)
			err = ws.WriteControl(websocket.PingMessage, nil, time.Now().Add(time.Minute))
		}
	}()

	fileCollectionController, err := file_collection.GetFileCollectionController(r.Context())
	if err != nil {
		return err
	}

	distanceChannel, err := archiveController.SearchForArchiveTo(query, method)
	if err != nil {
		return err
	}

	var rowCounter int
	for scannedDistance := range distanceChannel {
		if scannedDistance.Error != nil {
			return scannedDistance.Error
		}

		distance := scannedDistance.Value

		fileCollection, err := fileCollectionController.GetByID(distance.FileCollectionID)
		if err != nil {
			return err
		}
		var fileCount int64
		switch depth {
		case "deep":
			fileCount, err = fileCollectionController.CountFiles(fileCollection.FileCollectionID)
		case "shallow":
			fallthrough
		default:
			fileCount, err = fileCollectionController.ShallowCountFiles(fileCollection.FileCollectionID)
		}
		if err != nil {
			return err
		}
		subCollectionCount, err := fileCollectionController.CountSubCollections(fileCollection.FileCollectionID)
		if err != nil {
			return err
		}

		searchRow := ContainerSearchResult{
			FileCollectionID:         distance.FileCollectionID,
			Name:                     distance.ArchiveName,
			Sha1:                     distance.Sha1.String,
			FileCollectionInsertDate: fileCollection.InsertDate.String(),
			Count:                    fileCount,
			Packages:                 subCollectionCount,
			Distance:                 null.NewInt(distance.Distance, true),
		}

		// send searchRow via WebSocket
		if err := ws.WriteJSON(searchRow); err != nil {
			return errors.Wrapf(err, "error writing JSON websocket message")
		}
		rowCounter++
	}

	// send final row count
	if err := ws.WriteJSON(rowCounter); err != nil {
		return errors.Wrapf(err, "error writing row count")
	}

	log.Info().Str(zerolog.CallerFieldName, "webSocketContainerSearch").Int("rowCount", rowCounter).Msg("Finished sending initial rows")

	if auto {
		log.Info().Str(zerolog.CallerFieldName, "webSocketContainerSearch").Int("rowCount", rowCounter).Msg("Waiting for count requests")
		for messageType, payload, err := ws.ReadMessage(); messageType != websocket.CloseMessage; messageType, payload, err = ws.ReadMessage() {
			if err != nil {
				if websocket.IsCloseError(err, websocket.CloseNormalClosure) {
					break
				}

				return errors.Wrapf(err, "error reading auto message")
			}

			var requestResponse struct {
				Index       int   `json:"index"` // request
				ContainerID int64 `json:"id"`    // request
				Count       int64 `json:"count"` // response
			}
			if err := json.Unmarshal(payload, &requestResponse); err != nil {
				return errors.Wrapf(err, "error unmarshalling containerID")
			}

			// log.Debug().Str(zerolog.CallerFieldName, "webSocketContainerSearch").Int("rowCount", rowCount).Interface("request", requestResponse).Msg("Counting")

			requestResponse.Count, err = fileCollectionController.CountFiles(requestResponse.ContainerID)
			if err != nil {
				return errors.Wrapf(err, "error selecting detailed file count for %d", requestResponse.ContainerID)
			}

			if err := ws.WriteJSON(requestResponse); err != nil {
				return errors.Wrapf(err, "error writing JSON websocket message")
			}
		}
	} else {
		// wait for close
		for messageType, _, err := ws.ReadMessage(); messageType != websocket.CloseMessage; messageType, _, err = ws.ReadMessage() {
			if err != nil {
				if websocket.IsCloseError(err, websocket.CloseNormalClosure) {
					break
				}

				return errors.Wrapf(err, "error waiting for websocket close")
			}
		}
	}

	if err := ws.Close(); err != nil {
		return errors.Wrapf(err, "error closing websocket")
	}

	return nil
}
