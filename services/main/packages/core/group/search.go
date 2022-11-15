// Copyright (c) 2020 Wind River Systems, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at:
//       http://www.apache.org/licenses/LICENSE-2.0
// Unless required by applicable law or agreed to in writing, software  distributed
// under the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES
// OR CONDITIONS OF ANY KIND, either express or implied.

package group

import (
	"fmt"

	"github.com/pkg/errors"
)

type GroupSearchResult struct {
	ID       int64  `db:"id" json:"id"`
	Path     string `db:"path" json:"path"`
	Sub      int    `db:"sub" json:"sub"`
	Packages int    `db:"packages" json:"packages"`
	Distance int    `db:"distance" json:"distance"`
}

// HandleGroupSearch searches groups using the given query, and returns a list of groups
func (groups *GroupController) HandleGroupSearch(searchQuery string, method string) ([]GroupSearchResult, error) {
	var sql string
	var values []interface{}
	switch method {
	case "default":
		sql = "SELECT gc.id, build_group_path(gc.id) as path, " +
			"(SELECT COUNT(*) FROM group_container WHERE parent_id=gc.id) as sub, " +
			"(SELECT COUNT(*) FROM file_collection WHERE group_container_id=gc.id) as packages, " +
			"levenshtein(build_group_path(gc.id), $1, 10, 1, 100) as distance " +
			"FROM group_container gc ORDER BY distance"
		values = []interface{}{searchQuery}
	default:
		return nil, errors.New(fmt.Sprintf("method %s not recognized", method))
	}

	var results []GroupSearchResult
	rows, err := groups.DB.Queryx(sql, values...)
	if err != nil {
		return nil, errors.Wrapf(err, "error selecting from database")
	}
	defer rows.Close()

	for rows.Next() {
		var row GroupSearchResult
		rows.StructScan(&row)

		results = append(results, row)
	}

	return results, nil
}
