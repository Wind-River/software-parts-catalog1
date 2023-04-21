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
	"strings"

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
	METHOD_LEVENSHTEIN_LESS_EQUAL
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
	case METHOD_LEVENSHTEIN_LESS_EQUAL:
		return "levenshtein_less_equal"
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
	case "levenshtein_less_equal":
		return METHOD_LEVENSHTEIN_LESS_EQUAL
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

func (controller ArchiveController) SearchForArchiveAll(query string, method SearchMethod, insertCost int, deleteCost int, substituteCost int, maxDistance int) ([]ArchiveDistance, error) {
	rows, err := controller.searchForArchive(query, method, insertCost, deleteCost, substituteCost, maxDistance)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scan.ScanAll[ArchiveDistance](rows)
}

func (controller ArchiveController) SearchForArchiveTo(query string, method SearchMethod, insertCost int, deleteCost int, substituteCost int, maxDistance int) (chan scan.ScannedRow[ArchiveDistance], error) {
	rows, err := controller.searchForArchive(query, method, insertCost, deleteCost, substituteCost, maxDistance)
	if err != nil {
		return nil, err
	}

	return scan.ScanTo[ArchiveDistance](rows)
}

func (controller ArchiveController) searchForArchive(query string, method SearchMethod, insertCost int, deleteCost int, substituteCost int, maxDistance int) (*sqlx.Rows, error) {
	query = strings.ToLower(query)
	// We support several different kinds of comparisons.
	// like is effectively a sub-string search.
	// fast performs a levenshtein string comparison on the results of a sub-string search.
	// levenshtein performs a levenshtein string comparison on all archives in the database.
	selct := `SELECT a.sha256, a.part_id, archive_alias.name, ARRAY(SELECT name FROM archive_alias WHERE archive_alias.archive_sha256=a.sha256) AS names`
	from := `FROM archive a
	INNER JOIN archive_alias ON archive_alias.archive_sha256=a.sha256`
	var sql string
	var values []interface{}
	switch method {
	case METHOD_SUBSTRING:
		where := `WHERE LOWER(archive_alias.name) LIKE $1::TEXT ORDER BY archive_alias.name`
		sql = fmt.Sprintf("%s %s %s", selct, from, where)
		values = []interface{}{fmt.Sprintf("%%%s%%", query)}
	case METHOD_FAST_LEVENSHTEIN:
		selct = fmt.Sprintf(`%s,
		levenshtein(LOWER(archive_alias.name), $2, %d, %d, %d) AS distance`,
			selct, insertCost, deleteCost, substituteCost)
		where := `WHERE LOWER(archive_alias.name) LIKE $1::TEXT ORDER BY distance`
		sql = fmt.Sprintf("%s %s %s", selct, from, where)
		values = []interface{}{fmt.Sprintf("%%%s%%", query), query}
	case METHOD_LEVENSHTEIN:
		selct = fmt.Sprintf(`%s,
		levenshtein(LOWER(archive_alias.name), $1, %d, %d, %d) AS distance`,
			selct, insertCost, deleteCost, substituteCost)
		order := `ORDER BY distance`
		sql = fmt.Sprintf("%s %s %s", selct, from, order)
		values = []interface{}{query}
	case METHOD_LEVENSHTEIN_LESS_EQUAL:
		selct = fmt.Sprintf(`%s,
		levenshtein_less_equal(LOWER(archive_alias.name), $1, %d, %d, %d, %d) AS distance`,
			selct, insertCost, deleteCost, substituteCost, maxDistance)
		order := `ORDER BY distance`
		sql = fmt.Sprintf("%s %s %s", selct, from, order)
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
