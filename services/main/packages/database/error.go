// Copyright (c) 2020 Wind River Systems, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at:
//       http://www.apache.org/licenses/LICENSE-2.0
// Unless required by applicable law or agreed to in writing, software  distributed
// under the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES
// OR CONDITIONS OF ANY KIND, either express or implied.

package database

import (
	"database/sql"
	"fmt"
)

// ErrNoRows is a struct used to report the query, query values, and a message for a query that returned no results but was expected to
type ErrNoRows struct {
	Query   string
	Values  []interface{}
	Message string
}

// Error implements error.Error
func (e ErrNoRows) Error() string {
	message := e.Message
	if message == "" {
		message = sql.ErrNoRows.Error()
	}

	if e.Values == nil || len(e.Values) == 0 {
		return fmt.Sprintf("%s: \"%s\"", message, e.Query)
	}

	var values string
	for _, v := range e.Values {
		if values == "" {
			values = fmt.Sprintf("%#v", v)
		} else {
			values = fmt.Sprintf("%s %#v", values, v)
		}
	}

	return fmt.Sprintf("%s: \"%s\" with values: %s", message, e.Query, values)
}
