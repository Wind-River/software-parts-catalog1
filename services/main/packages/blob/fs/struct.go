// Copyright (c) 2020 Wind River Systems, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at:
//       http://www.apache.org/licenses/LICENSE-2.0
// Unless required by applicable law or agreed to in writing, software  distributed
// under the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES
// OR CONDITIONS OF ANY KIND, either express or implied.

package fs

import (
	"path/filepath"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"github.com/pkg/errors"
)

// Initialize BlobFileSystem

type BlobFileSystem struct {
	root string
	db   *sqlx.DB
}

func NewBlobFileSystem(root string, db *sqlx.DB) *BlobFileSystem {
	ret := new(BlobFileSystem)
	ret.root = root
	ret.db = db

	return ret
}

func CreateBlobFileSystem(root string) (*BlobFileSystem, error) {
	sqlite3FilePath := filepath.Join(root, ".db.sqlite3")
	db, err := sqlx.Open("sqlite3", "file:"+sqlite3FilePath)
	if err != nil {
		err = errors.Wrapf(err, "error creating %s", sqlite3FilePath)
		return nil, err
	}

	if _, err := db.Exec(`
		CREATE TABLE blob_metadata (
			sha256 BLOB PRIMARY KEY,
			sha1 BLOB NOT NULL,
			size BIGINT NOT NULL,
			mime TEXT
		)
	`); err != nil {
		err = errors.Wrapf(err, "error creating blob table")
		return nil, err
	}

	return NewBlobFileSystem(root, db), nil
}
