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
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"io"
	"os"

	"wrs/tk/packages/blob"

	"wrs/tk/packages/blob/file"

	"github.com/gabriel-vasile/mimetype"
	"github.com/pkg/errors"
)

type File struct {
	Sha256 [32]byte `db:"sha256"`
	Size   int64    `db:"file_size"`
	Md5    [16]byte `db:"md5"`
	Sha1   [20]byte `db:"sha1"`

	Aliases []string `db:"names"`

	LocalPath string
}

func NewFile(filePath string) (*File, error) {
	stat, err := os.Lstat(filePath)
	if err != nil {
		err = errors.Wrapf(err, "error stating %s", filePath)
		return nil, err
	}

	ret := new(File)
	ret.Aliases = []string{stat.Name()}
	ret.LocalPath = filePath
	ret.Size = stat.Size()

	if ret.Size == 0 || !stat.Mode().IsRegular() {
		copy(ret.Md5[:], md5.New().Sum(nil))       // empty md5
		copy(ret.Sha1[:], sha1.New().Sum(nil))     // empty sha1
		copy(ret.Sha256[:], sha256.New().Sum(nil)) // empty sha256
	} else {
		md5 := md5.New()
		sha1 := sha1.New()
		sha256 := sha256.New()

		f, err := os.Open(filePath)
		if err != nil {
			err = errors.Wrapf(err, "error opening file %s", filePath)
			return ret, err
		}
		defer f.Close()

		for {
			buf := make([]byte, 64)

			n, err := f.Read(buf)
			if err == io.EOF {
				break
			} else if err != nil {
				err = errors.Wrapf(err, "error chunking %s", filePath)
				return ret, err
			}

			buf = buf[:n]

			if _, err := md5.Write(buf); err != nil {
				err = errors.Wrapf(err, "error writing chunk to md5")
				return ret, err
			}
			if _, err := sha1.Write(buf); err != nil {
				err = errors.Wrapf(err, "error writing chunk to sha1")
				return ret, err
			}
			if _, err := sha256.Write(buf); err != nil {
				err = errors.Wrapf(err, "error writing chunk to sha256")
				return ret, err
			}
		}

		copy(ret.Md5[:], md5.Sum(nil))       // finish md5
		copy(ret.Sha1[:], sha1.Sum(nil))     // finish sha1
		copy(ret.Sha256[:], sha256.Sum(nil)) // finish sha256
	}

	return ret, nil
}

func StoreFile(blobStorage blob.Storage, f *File, filePath string) error {
	// Only store if file is a normal file greater than 0 bytes
	if f.Size < 1 {
		return nil
	}

	mimeType, err := mimetype.DetectFile(f.LocalPath)
	if err != nil {
		err = errors.Wrap(err, "error detecting mimetype")
		return err
	}

	r, err := os.Open(filePath)
	if err != nil {
		err = errors.Wrapf(err, "error opening file")
		return err
	}
	defer r.Close()

	if err := blobStorage.Store(r, &file.FileInfo{
		Sha256:   file.Sha256(f.Sha256),
		Sha1:     file.Sha1(f.Sha1),
		Size:     f.Size,
		MimeType: mimeType.String(),
	}); err != nil {
		return err
	}

	return nil
}
