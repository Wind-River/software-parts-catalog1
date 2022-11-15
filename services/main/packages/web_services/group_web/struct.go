// Copyright (c) 2020 Wind River Systems, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at:
//       http://www.apache.org/licenses/LICENSE-2.0
// Unless required by applicable law or agreed to in writing, software  distributed
// under the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES
// OR CONDITIONS OF ANY KIND, either express or implied.

package group_web

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"
	"wrs/tk/packages/core/archive"
	"wrs/tk/packages/core/file_collection"
	"wrs/tk/packages/core/group"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/log"
	"gopkg.in/guregu/null.v4"
)

type Group struct {
	Count       int               `json:"count"`
	Containers  []group.Container `json:"containers"`
	RequestTime time.Time         `json:"request_time"`
}

func HandleGroupQuery(w http.ResponseWriter, r *http.Request) {
	now := time.Now()
	queryValues := r.URL.Query()

	var gid int64
	var err error
	ids, hasID := queryValues["id"]
	if hasID {
		gid, err = strconv.ParseInt(ids[0], 10, 64)
		if err != nil {
			gid = 0
		}
	}

	groupController, err := group.GetGroupController(r)
	if err != nil {
		http.Error(w, "error getting group controller", 500)
		return
	}

	if gid == 0 {
		var path string
		paths, hasPath := queryValues["path"]
		if hasPath {
			path = paths[0]
		}

		gid, err = groupController.ParsePath(path)
		if err != nil {
			http.Error(w, "Error Parsing Path", 500)
			log.Error().Err(err).Str("path", path).Msg("error parsing group path")
			return
		}
	}

	containers, err := groupController.ListContainers(gid)
	if err != nil {
		http.Error(w, "Error Listing Containers", 500)
		log.Error().Err(err).Int64("group_id", gid).Msg("error listing containers for group")
		return
	}

	ret := Group{}
	ret.Containers = containers
	ret.Count = len(containers)
	ret.RequestTime = now

	if err := json.NewEncoder(w).Encode(ret); err != nil {
		http.Error(w, "Error encoding JSON response", 500)
		log.Error().Err(err).Msg("error encoding JSON response")
		return
	}
}

type ShowPackage struct {
	Name  string      `db:"name" json:"name"`
	Count int64       `db:"count" json:"count"` // number of files directly contained by package
	Date  time.Time   `db:"insert_date" json:"date"`
	Sha1  null.String `db:"checksum_sha1" json:"sha1"`
}

type ShowGroup struct {
	ID       int64  `db:"id" json:"id"`
	Path     string `db:"path" json:"path"`
	Sub      int    `db:"sub" json:"sub"`
	Packages int    `db:"packages" json:"packages"`
}

// HandleGroupGet returns information on a given group.
// This information includes information on the group, any sub-groups ,and packages belinging to the group.
// The URLParam groupID is the group table's id column.
func HandleGroupGet(w http.ResponseWriter, r *http.Request) {
	idString := chi.URLParam(r, "groupID")

	if idString == "" {
		http.Error(w, "id missing from query string", 400)
		return
	}

	id, err := strconv.ParseInt(idString, 10, 64)
	if err != nil {
		http.Error(w, "unable to parse given id", 400)
		return
	}

	groupController, err := group.GetGroupController(r)
	if err != nil {
		http.Error(w, "error getting group controller", 500)
		return
	}
	fileCollectionController, err := file_collection.GetFileCollectionController(r.Context())
	if err != nil {
		http.Error(w, "error getting file collection controller", 500)
		return
	}
	archiveController, err := archive.GetArchiveController(r.Context())
	if err != nil {
		http.Error(w, "error getting archive controller", 500)
		return
	}

	group, err := groupController.GetByID(id)
	if err != nil {
		log.Error().Err(err).Int64("group_id", id).Msg("error getting grouy by id")
		http.Error(w, "error getting group by id", 500)
		return
	}

	ret := struct {
		Group    ShowGroup     `json:"group"`
		Subs     []ShowGroup   `json:"subs"`
		Packages []ShowPackage `json:"packages"`
	}{
		Subs:     make([]ShowGroup, 0),
		Packages: make([]ShowPackage, 0),
	}

	ret.Group.ID = group.GroupID

	if path, err := groupController.BuildGroupPath(group.GroupID); err != nil {
		http.Error(w, "error building group path", 500)
		return
	} else {
		ret.Group.Path = path
	}

	if groupCount, collectionCount, err := groupController.CountRelations(group.GroupID); err != nil {
		http.Error(w, "error counting group relations", 500)
		return
	} else {
		ret.Group.Sub = groupCount
		ret.Group.Packages = collectionCount
	}

	subs, err := groupController.GetByParentID(group.GroupID)
	if err != nil {
		http.Error(w, "error getting sub-groups", 500)
		return
	}
	for _, sub := range subs {
		path, err := groupController.BuildGroupPath(sub.GroupID)
		if err != nil {
			http.Error(w, "error building group path", 500)
			return
		}

		groupCount, packageCount, err := groupController.CountRelations(sub.GroupID)
		if err != nil {
			http.Error(w, "error counting sub group relations", 500)
			return
		}

		var tmp ShowGroup
		tmp.ID = sub.GroupID
		tmp.Path = path
		tmp.Sub = groupCount
		tmp.Packages = packageCount

		ret.Subs = append(ret.Subs, tmp)
	}

	containers, err := groupController.ListContainers(group.GroupID)
	if err != nil {
		http.Error(w, "error listing containers", 500)
		return
	}

	for _, container := range containers {
		count, err := fileCollectionController.CountFiles(container.ID)
		if err != nil {
			http.Error(w, "error counting file_collection files", 500)
			return
		}
		archives, err := archiveController.GetByFileCollection(container.ID)
		if err != nil {
			http.Error(w, "error getting archives of file_collection", 500)
			return
		}

		archive := archives[0]

		var tmp ShowPackage
		tmp.Name = container.Names[0]
		tmp.Count = count
		tmp.Date = archive.InsertDate
		if archive.Sha1.Valid {
			tmp.Sha1.SetValid(archive.Sha1.String)
		}

		ret.Packages = append(ret.Packages, tmp)
	}

	if err = json.NewEncoder(w).Encode(ret); err != nil {
		http.Error(w, "error encoding json response", 500)
		log.Error().Err(err).Msg("error encoding json response")
		return
	}
}
