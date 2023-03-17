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

	"wrs/tk/packages/core/part"

	"github.com/pkg/errors"
)

func (controller ArchiveController) GetBy(sha256 []byte, sha1 []byte, name string) (*Archive, error) {
	var ret *Archive = nil
	var err error

	if sha256 != nil { // try fetching by sha256
		ret, err = controller.GetBySha256(sha256)
		if err != nil && err != ErrNotFound {
			return nil, err
		}
	}
	if sha1 != nil { // try fetching by sha1
		ret, err = controller.GetBySha1(sha1)
		if err != nil && err != ErrNotFound {
			return nil, err
		}
	}
	if name != "" { // try fetching by archive name
		ret, err = controller.GetByName(name)
		if err != nil && err != ErrNotFound {
			return nil, err
		}
	}

	return ret, ErrNotFound
}

func (controller ArchiveController) GetBySha256(sha256 []byte) (*Archive, error) {
	if sha256 == nil || len(sha256) != 32 {
		return nil, ErrNotFound
	}

	ret := new(Archive)
	if err := controller.DB.QueryRowx("SELECT archive.* FROM archive WHERE sha256=$1",
		sha256).StructScan(ret); err == sql.ErrNoRows {
		return nil, ErrNotFound
	} else if err != nil {
		return nil, errors.Wrapf(err, "error getting archive by sha256:%x", sha256)
	}

	rows, err := controller.DB.Query("SELECT name FROM archive_alias WHERE archive_sha256=$1",
		ret.Sha256)
	if err != nil && err != sql.ErrNoRows {
		return nil, errors.Wrapf(err, "error selecting archive_aliases")
	}
	defer rows.Close()

	for rows.Next() {
		if ret.Aliases == nil {
			ret.Aliases = make([]string, 0)
		}

		var tmp string
		if err := rows.Scan(&tmp); err != nil {
			return nil, errors.Wrapf(err, "error scanning archive_alias")
		}

		ret.Aliases = append(ret.Aliases, tmp)
	}

	return ret, nil
}

func (controller ArchiveController) GetBySha1(sha1 []byte) (*Archive, error) {
	if sha1 == nil || len(sha1) != 20 {
		return nil, ErrNotFound
	}

	ret := new(Archive)
	if err := controller.DB.QueryRowx("SELECT archive.* FROM archive WHERE sha1=$1",
		sha1).StructScan(ret); err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}

		rows, err := controller.DB.Query("SELECT name FROM archive_alias WHERE archive_sha256=$1",
			ret.Sha256)
		if err != nil && err != sql.ErrNoRows {
			return nil, errors.Wrapf(err, "error selecting archive_aliases")
		}
		defer rows.Close()

		for rows.Next() {
			if ret.Aliases == nil {
				ret.Aliases = make([]string, 0)
			}

			var tmp string
			if err := rows.Scan(&tmp); err != nil {
				return nil, errors.Wrapf(err, "error scanning archive_alias")
			}

			ret.Aliases = append(ret.Aliases, tmp)
		}

		return nil, err
	}

	return ret, nil
}

func (controller ArchiveController) GetByName(name string) (*Archive, error) {
	if name == "" {
		return nil, ErrNotFound
	}

	ret := new(Archive)
	if err := controller.DB.QueryRowx("SELECT archive.* "+
		"FROM archive "+
		"INNER JOIN archive_alias ON archive_alias.archive_sha256=archive.sha256 "+
		"WHERE archive_alias.name=$1",
		name).StructScan(ret); err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}

		rows, err := controller.DB.Query("SELECT name FROM archive_alias WHERE archive_sha256=$1",
			ret.Sha256)
		if err != nil && err != sql.ErrNoRows {
			return nil, errors.Wrapf(err, "error selecting archive_aliases")
		}
		defer rows.Close()

		for rows.Next() {
			if ret.Aliases == nil {
				ret.Aliases = make([]string, 0)
			}

			var tmp string
			if err := rows.Scan(&tmp); err != nil {
				return nil, errors.Wrapf(err, "error scanning archive_alias")
			}

			ret.Aliases = append(ret.Aliases, tmp)
		}

		return nil, err
	}

	return ret, nil
}

func (controller ArchiveController) GetByPart(partID part.ID) ([]Archive, error) {
	ret := make([]Archive, 0)

	rows, err := controller.DB.Queryx(`SELECT archive.*
	FROM archive
	WHERE archive.part_id=$1
	`, partID)
	if err != nil {
		return nil, errors.Wrapf(err, "error selecting archives by part_id")
	}
	defer rows.Close()

	for rows.Next() {
		var arch Archive
		if err := rows.StructScan(&arch); err != nil {
			return nil, errors.Wrapf(err, "error scanning archives by part_id")
		}

		ret = append(ret, arch)
	}
	rows.Close()

	for i, arch := range ret {
		rows, err := controller.DB.Query("SELECT name FROM archive_alias WHERE archive_sha256=$1",
			arch.Sha256)
		if err != nil && err != sql.ErrNoRows {
			return nil, errors.Wrapf(err, "error selecting archive_aliases")
		}
		defer rows.Close()

		for rows.Next() {
			if ret[i].Aliases == nil {
				ret[i].Aliases = make([]string, 0)
			}

			var tmp string
			if err := rows.Scan(&tmp); err != nil {
				return nil, errors.Wrapf(err, "error scanning archive_alias")
			}

			ret[i].Aliases = append(ret[i].Aliases, tmp)
		}
	}

	return ret, nil
}
