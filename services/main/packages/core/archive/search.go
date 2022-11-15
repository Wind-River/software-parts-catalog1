// Copyright (c) 2020 Wind River Systems, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at:
//       http://www.apache.org/licenses/LICENSE-2.0
// Unless required by applicable law or agreed to in writing, software  distributed
// under the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES
// OR CONDITIONS OF ANY KIND, either express or implied.

package archive

import (
	"database/sql"
	"fmt"

	scan "wrs/tk/packages/generics/rows"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type SearchMethod int

const (
	METHOD_UNKNOWN SearchMethod = iota
	METHOD_SUBSTRING
	METHOD_LEVENSHTEIN
	METHOD_FAST_LEVENSHTEIN
)

func SearchMethodString(m SearchMethod) string {
	switch m {
	case METHOD_UNKNOWN:
		return "unknown"
	case METHOD_SUBSTRING:
		return "like"
	case METHOD_LEVENSHTEIN:
		return "levenshtein"
	case METHOD_FAST_LEVENSHTEIN:
		return "fast"
	}

	return fmt.Sprintf("unrecognized{%d}", m)
}

func ParseMethod(key string) SearchMethod {
	switch key {
	case "like":
		return METHOD_SUBSTRING
	case "levenshtein":
		return METHOD_LEVENSHTEIN
	case "fast":
		return METHOD_FAST_LEVENSHTEIN
	default:
		return METHOD_UNKNOWN
	}
}

type ArchiveDistance struct {
	ArchiveID        int64          `db:"archive_id"`
	ArchiveName      string         `db:"name"`
	Sha1             sql.NullString `db:"checksum_sha1"`
	FileCollectionID int64          `db:"file_collection_id"`
	Distance         int64          `db:"distance"`
}

func (controller ArchiveController) SearchForArchiveAll(query string, method SearchMethod) ([]ArchiveDistance, error) {
	rows, err := controller.searchForArchive(query, method)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scan.ScanAll[ArchiveDistance](rows)
}

func (controller ArchiveController) SearchForArchiveTo(query string, method SearchMethod) (chan scan.ScannedRow[ArchiveDistance], error) {
	rows, err := controller.searchForArchive(query, method)
	if err != nil {
		return nil, err
	}

	return scan.ScanTo[ArchiveDistance](rows)
}

func (controller ArchiveController) searchForArchive(query string, method SearchMethod) (*sqlx.Rows, error) {
	// We support several different kinds of comparisons.
	// like is effectively a sub-string search.
	// fast performs a levenshtein string comparison on the results of a sub-string search.
	// levenshtein performs a levenshtein string comparison on all archives in the database.
	var sql string
	var values []interface{}
	switch method {
	case METHOD_SUBSTRING:
		sql = fmt.Sprintf("SELECT a.id as achive_id, a.name, a.checksum_sha1, c.id as file_collection_id, " +
			"FROM file_collection c INNER JOIN archive a ON a.file_collection_id=c.id WHERE a.name LIKE $1::text ORDER BY a.name")
		values = []interface{}{fmt.Sprintf("%%%s%%", query)}
	case METHOD_FAST_LEVENSHTEIN:
		sql = fmt.Sprintf("SELECT a.id as archive_id, a.name, a.checksum_sha1, c.id as file_collection_id, " +
			"levenshtein(a.name, $2, 10, 1, 100) as distance " +
			"FROM file_collection c INNER JOIN archive a ON a.file_collection_id=c.id WHERE a.name LIKE $1::text ORDER BY distance")
		values = []interface{}{fmt.Sprintf("%%%s%%", query), query}
	case METHOD_LEVENSHTEIN:
		sql = fmt.Sprintf("SELECT a.id as archive_id, a.name, a.checksum_sha1, c.id as file_collection_id, " +
			"levenshtein(a.name, $1, 10, 1, 100) as distance " +
			"FROM file_collection c INNER JOIN archive a ON a.file_collection_id=c.id ORDER BY distance")
		values = []interface{}{query}
		for _, v := range values {
			fmt.Printf("v: %s\n", v)
		}
	default:
		msg := fmt.Sprintf("Method %s not recognized", SearchMethodString(method))
		return nil, errors.New(msg)
	}

	log.Debug().Str(zerolog.CallerFieldName, "webSocketContainerSearch").Str("sql", sql).Interface("values", values).Send()

	ret, err := controller.DB.Queryx(sql, values...)
	if err != nil {
		return nil, errors.Wrapf(err, "error selecting rows")
	}

	return ret, nil
}
