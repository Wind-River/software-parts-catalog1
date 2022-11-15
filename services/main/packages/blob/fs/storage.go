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
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"io"
	"os"
	"path/filepath"
	"wrs/tk/packages/blob/file"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

// Implentation of the blob.Storage interface

func (fs BlobFileSystem) Store(data io.Reader, metadata *file.FileInfo) error {
	fullPath := fs.root
	for _, v := range metadata.Sha256 {
		fullPath = filepath.Join(fullPath, hex.EncodeToString([]byte{v}))
	}
	directoryPath := filepath.Dir(fullPath)

	var info file.FileInfo
	var exists bool
	if err := fs.db.QueryRowx("SELECT * FROM blob_metadata WHERE sha256=$1", info.Sha256).StructScan(&info); err != nil && err != sql.ErrNoRows {
		err = errors.Wrapf(err, "error checking if blob exists")
		return err
	} else if err == nil {
		exists = true
	}

	if exists {
		if info.Sha256.Hex() == metadata.Sha256.Hex() &&
			info.Sha1.Hex() == metadata.Sha1.Hex() &&
			info.Size == metadata.Size {
			// Verify local file
			if fs.hasLocally(*metadata) {
				return nil
			}
		} // else overwrite existing
	}

	if err := os.MkdirAll(directoryPath, 0775); err != nil {
		err = errors.Wrapf(err, "error mkdir -p %s", directoryPath)
		return err
	}

	f, err := os.Create(fullPath)
	if err != nil {
		err = errors.Wrapf(err, "error opening %s", fullPath)
		return err
	}
	defer f.Close()

	g := gzip.NewWriter(f)
	defer g.Close()

	if _, err := io.Copy(g, data); err != nil {
		err = errors.Wrapf(err, "error copying data")
		return err
	}

	// upsert
	if _, err := fs.db.Exec("INSERT INTO blob_metadata(size, mime, sha256, sha1) VALUES ($1, $2, $3, $4) ON CONFLICT (sha256) DO UPDATE SET "+
		"sha1=EXCLUDED.sha1, size=EXCLUDED.size, mime=EXCLUDED.mime",
		metadata.Size, metadata.MimeType, metadata.Sha256, metadata.Sha1); err != nil {
		err = errors.Wrapf(err, "error inserting metadata")
		return err
	}

	return nil
}

func (fs BlobFileSystem) hasLocally(info file.FileInfo) bool {
	fullPath := fs.root
	for _, v := range info.Sha256 {
		fullPath = filepath.Join(fullPath, hex.EncodeToString([]byte{v}))
	}

	f, err := os.Open(fullPath)
	if err != nil {
		return false
	}
	defer f.Close()

	g, err := gzip.NewReader(f)
	if err != nil {
		return false
	}
	defer g.Close()

	hasher := sha256.New()
	written, err := io.Copy(hasher, g)
	if err != nil {
		return false
	}

	if bytes.Compare(hasher.Sum(nil), info.Sha256.Bytes()) != 0 {
		return false
	}

	if written < int64(info.Size) {
		return false
	}

	return true
}

func (fs BlobFileSystem) Retrieve(hash file.Sha256) (*file.File, error) {
	var info file.FileInfo
	if err := fs.db.QueryRowx("SELECT * FROM blob_metadata WHERE sha256=$1", hash).StructScan(&info); err != nil {
		err = errors.Wrapf(err, "error selecting %x", hash)
		return nil, err
	}
	log.Trace().Interface("FileInfo", info).Msg("Scanned Struct")

	fullPath := fs.root
	for _, v := range info.Sha256 {
		fullPath = filepath.Join(fullPath, hex.EncodeToString([]byte{v}))
	}

	f, err := os.OpenFile(fullPath, os.O_RDONLY, 0777)
	if err != nil {
		err = errors.Wrapf(err, "error opening file %s", fullPath)
		return nil, err
	}

	return &file.File{
		FileInfo:       info,
		ReadSeekCloser: f,
	}, nil
}

func (fs BlobFileSystem) ListAll() ([]file.FileInfo, error) {
	rows, err := fs.db.Queryx("SELECT * FROM blob_metadata")
	if err != nil {
		err = errors.Wrapf(err, "error selecting all")
		return nil, err
	}
	defer rows.Close()

	ret := make([]file.FileInfo, 0)
	for rows.Next() {
		var tmp file.FileInfo
		if err := rows.StructScan(&tmp); err != nil {
			err = errors.Wrapf(err, "error scanning")
			return ret, err
		}

		ret = append(ret, tmp)
	}

	return ret, nil
}

func (fs BlobFileSystem) StreamAll() (chan file.FileInfo, error) {
	rows, err := fs.db.Queryx("SELECT * FROM blob_metadata")
	if err != nil {
		err = errors.Wrapf(err, "error selecting all")
		return nil, err
	}

	ch := make(chan file.FileInfo)

	go func(rows *sqlx.Rows, ch chan file.FileInfo) error {
		defer rows.Close()
		defer close(ch)

		for rows.Next() {
			var tmp file.FileInfo
			if err := rows.StructScan(&tmp); err != nil {
				err = errors.Wrapf(err, "error scanning")
				return err
			}

			ch <- tmp
		}

		return nil
	}(rows, ch)

	return ch, nil
}
