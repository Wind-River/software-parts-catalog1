// Copyright (c) 2020 Wind River Systems, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at:
//       http://www.apache.org/licenses/LICENSE-2.0
// Unless required by applicable law or agreed to in writing, software  distributed
// under the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES
// OR CONDITIONS OF ANY KIND, either express or implied.

package part

import (
	"database/sql"

	"github.com/pkg/errors"
)

// GetAliases lists every alias of the give part
// slice will be nil if no aliases for this part exists
func (controller PartController) GetAliases(partID ID) ([]string, error) {
	rows, err := controller.DB.Queryx("SELECT alias FROM part_alias WHERE part_id=$1", partID)
	if err == sql.ErrNoRows {
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	defer rows.Close()

	ret := make([]string, 0)
	for rows.Next() {
		var tmp string
		if err := rows.Scan(&tmp); err != nil {
			return nil, errors.Wrapf(err, "error scanning alias")
		}

		ret = append(ret, tmp)
	}

	return ret, nil
}
