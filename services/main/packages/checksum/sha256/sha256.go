// Copyright (c) 2020 Wind River Systems, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at:
//       http://www.apache.org/licenses/LICENSE-2.0
// Unless required by applicable law or agreed to in writing, software  distributed
// under the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES
// OR CONDITIONS OF ANY KIND, either express or implied.

package sha256

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"wrs/tk/packages/checksum"

	"github.com/pkg/errors"
)

func Sha256OfFilePath(filePath string) (string, error) {
	if filePath == "" {
		return EmptySha256(), &checksum.EmptyChecksum{Algorithm: "sha256"}
	}

	f, err := os.Open(filePath)
	if err != nil {
		return "", errors.Wrapf(err, "error opening %s for SHA256", filePath)
	}
	defer f.Close()

	return FileToSha256(f)
}

func EmptySha256() string {
	hasher := sha256.New()
	hash := fmt.Sprintf("%x", hasher.Sum(nil))

	return hash
}

func FileToSha256(input *os.File) (string, error) {
	hasher := sha256.New()
	if _, err := io.Copy(hasher, input); err != nil {
		return "", errors.Wrap(err, "error writing file to SHA256")
	}
	hash := fmt.Sprintf("%x", hasher.Sum(nil))

	return hash, nil
}

func BytesToSha256(input []byte) (string, error) {
	hasher := sha256.New()

	buf := bytes.NewBuffer(input)
	buf.WriteTo(hasher)

	hash := fmt.Sprintf("%x", hasher.Sum(nil))

	return hash, nil
}

// Sha256 calculates the sha256 of the given file and returns the sha256 as a fixed-size byte array.
func RawSha256(filePath string) ([32]byte, error) {
	if filePath == "" {
		return EmptyRawSha256(), &checksum.EmptyChecksum{Algorithm: "sha256"}
	}

	f, err := os.Open(filePath)
	if err != nil {
		return [32]byte{}, errors.Wrap(err, "error opening file")
	}
	defer f.Close()

	hasher := sha256.New()
	if _, err := io.Copy(hasher, f); err != nil {
		return [32]byte{}, errors.Wrap(err, "error copying to hasher")
	}

	var ret [32]byte
	copy(ret[:], hasher.Sum(nil))

	return ret, nil
}

// EmptySha256 returns the fixed-size byte array form of the sha256 of nil data.
func EmptyRawSha256() [32]byte {
	return sha256.Sum256(nil)
}
