// Copyright (c) 2020 Wind River Systems, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at:
//       http://www.apache.org/licenses/LICENSE-2.0
// Unless required by applicable law or agreed to in writing, software  distributed
// under the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES
// OR CONDITIONS OF ANY KIND, either express or implied.

package file

import (
	"bytes"
	"crypto/sha1"
	"crypto/sha256"
	"database/sql/driver"
	"encoding/hex"
	"io"

	"github.com/pkg/errors"
)

type FileInfo struct {
	Sha256   Sha256 `db:"sha256" json:"sha256"`
	Sha1     Sha1   `db:"sha1" json:"sha1"`
	Size     int64  `db:"size" json:"size"`
	MimeType string `db:"mime" json:"mime"`
}

type File struct {
	FileInfo
	io.ReadSeekCloser
}

type Sha256 [32]byte

func ParseSha256(s string) (*Sha256, error) {
	var ret Sha256
	if _, err := hex.Decode(ret[:], []byte(s)); err != nil {
		err = errors.Wrapf(err, "error parsing %s", s)
		return nil, err
	}

	return &ret, nil
}

func (h Sha256) IsValid() bool {
	return bytes.Compare(h.Bytes(), make([]byte, 32, 32)) != 0
}

func (h Sha256) IsEmpty() bool {
	return bytes.Compare(h.Bytes(), emptySha256().Bytes()) == 0
}

func emptySha256() *Sha256 {
	var h Sha256 = sha256.Sum256(nil)
	return &h
}

func (h *Sha256) Scan(value interface{}) error {
	var err error
	if value == nil {
		h = emptySha256()
		return nil
	}

	var sv driver.Value
	sv, err = driver.String.ConvertValue(value)
	if err == nil {
		if v, ok := sv.([]byte); ok {
			if _, err := hex.Decode(h[:], v); err != nil {
				copy(h[:], v)
			}
		}
	}

	return errors.Wrapf(err, "failed to scan Sha256")
}

func (h Sha256) Value() (driver.Value, error) {
	return h.Bytes(), nil
}

func (h Sha256) Bytes() []byte {
	return h[:]
}

func (h Sha256) Array() [32]byte {
	return h
}

func (h Sha256) Hex() string {
	return hex.EncodeToString(h.Bytes())
}

type Sha1 [20]byte

func ParseSha1(s string) (*Sha1, error) {
	var ret Sha1
	if _, err := hex.Decode(ret[:], []byte(s)); err != nil {
		err = errors.Wrapf(err, "error parsing %s", s)
		return nil, err
	}

	return &ret, nil
}
func (h Sha1) IsValid() bool {
	return bytes.Compare(h.Bytes(), make([]byte, 20, 20)) != 0
}

func (h Sha1) IsEmpty() bool {
	return bytes.Compare(h.Bytes(), emptySha1().Bytes()) == 0
}

func emptySha1() *Sha1 {
	var h Sha1 = sha1.Sum(nil)
	return &h
}

func (h *Sha1) Scan(value interface{}) error {
	var err error
	if value == nil {
		h = emptySha1()
		return nil
	}

	var sv driver.Value
	sv, err = driver.String.ConvertValue(value)
	if err == nil {
		if v, ok := sv.([]byte); ok {
			if _, err := hex.Decode(h[:], v); err != nil {
				copy(h[:], v)
			}
		}
	}

	return errors.Wrapf(err, "failed to scan Sha1")
}

func (h Sha1) Value() (driver.Value, error) {
	return h.Bytes(), nil
}

func (h Sha1) Bytes() []byte {
	return h[:]
}

func (h Sha1) Array() [20]byte {
	return h
}

func (h Sha1) Hex() string {
	return hex.EncodeToString(h.Bytes())
}
