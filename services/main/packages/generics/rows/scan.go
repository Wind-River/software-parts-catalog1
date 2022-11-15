// Copyright (c) 2020 Wind River Systems, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at:
//       http://www.apache.org/licenses/LICENSE-2.0
// Unless required by applicable law or agreed to in writing, software  distributed
// under the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES
// OR CONDITIONS OF ANY KIND, either express or implied.

package rows

import (
	"github.com/jmoiron/sqlx"
)

// ScanAll returns an array of generic type V from scanning all rows
// If an error occurs when scanning any of the rows, execution stops, and the rows scanned so far and the error are returned
func ScanAll[V any](rows *sqlx.Rows) ([]V, error) {
	ret := make([]V, 0)

	for rows.Next() {
		var tmp V
		if err := rows.StructScan(&tmp); err != nil {
			return ret, err
		}
		ret = append(ret, tmp)
	}

	return ret, nil
}

type ScannedRow[V any] struct {
	Value *V
	Error error
}

// ScanTo streams scanned rows of generic type V to a go channel
// If an error occurs when scanning, the error is included in the channel message, and execution continues
func ScanTo[V any](rows *sqlx.Rows) (chan ScannedRow[V], error) {
	ret := make(chan ScannedRow[V])
	go scanTo(rows, ret)
	return ret, nil
}

func scanTo[V any](rows *sqlx.Rows, returnChannel chan ScannedRow[V]) {
	for rows.Next() {
		var tmp V
		err := rows.StructScan(&tmp)
		returnChannel <- ScannedRow[V]{
			Value: &tmp,
			Error: err,
		}
	}
	rows.Close()

	close(returnChannel)
}
