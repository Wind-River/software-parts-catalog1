// Copyright (c) 2020 Wind River Systems, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at:
//       http://www.apache.org/licenses/LICENSE-2.0
// Unless required by applicable law or agreed to in writing, software  distributed
// under the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES
// OR CONDITIONS OF ANY KIND, either express or implied.

package sha1

import (
	"bytes"
	"crypto/sha1"
	"fmt"
	"io"
	"os"
	"wrs/tk/packages/checksum"

	"github.com/pkg/errors"
)

func Sha1OfFilePath(filePath string) (string, error) {
	if filePath == "" {
		return EmptySha1(), &checksum.EmptyChecksum{Algorithm: "sha1"}
	}

	f, err := os.Open(filePath)
	if err != nil {
		return "", errors.Wrapf(err, "error opening %s for SHA1", filePath)
	}
	defer f.Close()

	return FileToSha1(f)
}

func EmptySha1() string {
	hasher := sha1.New()
	hash := fmt.Sprintf("%x", hasher.Sum(nil))

	return hash
}

func FileToSha1(input *os.File) (string, error) {
	hasher := sha1.New()
	if _, err := io.Copy(hasher, input); err != nil {
		return "", errors.Wrap(err, "error writing file to SHA1 hasher")
	}
	hash := fmt.Sprintf("%x", hasher.Sum(nil))

	return hash, nil
}

func BytesToSha1(input []byte) (string, error) {
	hasher := sha1.New()

	buf := bytes.NewBuffer(input)
	buf.WriteTo(hasher)

	hash := fmt.Sprintf("%x", hasher.Sum(nil))

	return hash, nil
}
