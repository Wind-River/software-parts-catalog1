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
	"fmt"

	scan "wrs/tk/packages/generics/rows"

	// "wrs/tk/packages/generics/slice"
	"github.com/jackc/pgtype"

	"github.com/google/uuid"
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
	// Sha256 [32]byte `db:"sha256"`
	Sha256         []byte           `db:"sha256"`
	PartID         uuid.UUID        `db:"part_id"`
	ArchiveAliases pgtype.TextArray `db:"names"` // TODO is this necessary here?
	MatchedName    string           `db:"name"`
	Distance       int64            `db:"distance"`
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
		sql = `SELECT a.sha256, a.part_id, archive_alias.name, ARRAY(SELECT name FROM archive_alias WHERE archive_alias.archive_sha256=a.sha256) AS names
		FROM archive a
		INNER JOIN archive_alias ON archive_alias.archive_sha256=a.sha256
		WHERE archive_alias.name LIKE $1::TEXT ORDER BY archive_alias.name`
		values = []interface{}{fmt.Sprintf("%%%s%%", query)}
	case METHOD_FAST_LEVENSHTEIN:
		sql = `SELECT a.sha256, a.part_id, archive_alias.name, ARRAY(SELECT name FROM archive_alias WHERE archive_alias.archive_sha256=a.sha256) AS names,
		levenshtein(archive_alias.name, $2, 10, 1, 100) AS distance
		FROM archive a
		INNER JOIN archive_alias ON archive_alias.archive_sha256=a.sha256
		WHERE archive_alias.name LIKE $1::TEXT ORDER BY distance`
		values = []interface{}{fmt.Sprintf("%%%s%%", query), query}
	case METHOD_LEVENSHTEIN:
		sql = `SELECT a.sha256, a.part_id, archive_alias.name, ARRAY(SELECT name FROM archive_alias WHERE archive_alias.archive_sha256=a.sha256) AS names,
		levenshtein(archive_alias.name, $1, 10, 1, 100) AS distance
		FROM archive a
		INNER JOIN archive_alias ON archive_alias.archive_sha256=a.sha256
		ORDER BY distance`
		values = []interface{}{query}
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
