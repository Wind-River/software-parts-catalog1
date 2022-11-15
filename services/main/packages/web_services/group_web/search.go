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

	"wrs/tk/packages/core/group"

	"github.com/rs/zerolog/log"
)

// HandleGroupSearch searches groups using the given query, and returns a list of groups
func HandleGroupSearch(w http.ResponseWriter, r *http.Request) {
	queryValues := r.URL.Query()

	_, hasQuery := queryValues["query"]
	if !hasQuery {
		http.Error(w, "Missing url parameter: 'query'", 400)
		return
	}
	searchQuery := queryValues.Get("query")

	var method string
	_, hasMethod := queryValues["method"]
	if !hasMethod {
		method = "default"
	} else {
		method = queryValues.Get("method")
	}

	groupController, err := group.GetGroupController(r)
	if err != nil {
		http.Error(w, "error getting group controller", 500)
		return
	}

	results, err := groupController.HandleGroupSearch(searchQuery, method)
	if err != nil {
		http.Error(w, "error searching for group", 500)
		return
	}

	if err = json.NewEncoder(w).Encode(results); err != nil {
		http.Error(w, "error encoding json response", 500)
		log.Error().Err(err).Msg("error encoding json response")
		return
	}
}
