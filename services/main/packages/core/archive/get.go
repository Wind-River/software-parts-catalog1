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

	"github.com/pkg/errors"
)

func (controller ArchiveController) GetBy(sha256 string, sha1 string, name string) (*Archive, error) {
	var ret *Archive = nil
	var err error

	if sha256 != "" { // try fetching by sha256
		ret, err = controller.GetBySha256(sha256)
		if err != nil && err != ErrNotFound {
			return nil, err
		}
	}
	if sha1 != "" { // try fetching by sha1
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

func (controller ArchiveController) GetBySha256(sha256 string) (*Archive, error) {
	if sha256 == "" || len(sha256) != 64 {
		return nil, ErrNotFound
	}

	var ret Archive
	if err := controller.DB.QueryRowx("SELECT * FROM archive WHERE checksum_sha256=$1",
		sha256).StructScan(&ret); err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}

		return nil, err
	}

	return &ret, nil
}

func (controller ArchiveController) GetBySha1(sha1 string) (*Archive, error) {
	if sha1 == "" || len(sha1) != 40 {
		return nil, ErrNotFound
	}

	var ret Archive
	if err := controller.DB.QueryRowx("SELECT * FROM archive WHERE checksum_sha1=$1",
		sha1).StructScan(&ret); err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}

		return nil, err
	}

	return &ret, nil
}

func (controller ArchiveController) GetByName(name string) (*Archive, error) {
	if name == "" {
		return nil, ErrNotFound
	}

	var ret Archive
	if err := controller.DB.QueryRowx("SELECT * FROM archive WHERE name=$1",
		name).StructScan(&ret); err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}

		return nil, err
	}

	return &ret, nil
}

func (controller ArchiveController) GetByFileCollection(fileCollectionID int64) ([]Archive, error) {
	if fileCollectionID <= 0 {
		return nil, ErrNotFound
	}

	ret := make([]Archive, 0)

	rows, err := controller.DB.Queryx("SELECT * FROM archive WHERE file_collection_id=$1",
		fileCollectionID)
	if err != nil {
		return nil, errors.Wrapf(err, "error selecting archives by file_collection_id")
	}
	defer rows.Close()

	for rows.Next() {
		var arch Archive
		if err := rows.StructScan(&arch); err != nil {
			return nil, errors.Wrapf(err, "error scanning archives by file_collection_id")
		}

		ret = append(ret, arch)
	}

	return ret, nil
}

func (controller ArchiveController) GetByID(archiveID int64) (*Archive, error) {
	if archiveID <= 0 {
		return nil, ErrNotFound
	}

	var ret Archive

	if err := controller.DB.QueryRowx("SELECT * FROM archive WHERE id=$1",
		archiveID).StructScan(&ret); err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}

		return nil, errors.Wrapf(err, "error selecting archive by id")
	}

	return &ret, nil
}
