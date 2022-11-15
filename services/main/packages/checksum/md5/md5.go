// Copyright (c) 2020 Wind River Systems, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at:
//       http://www.apache.org/licenses/LICENSE-2.0
// Unless required by applicable law or agreed to in writing, software  distributed
// under the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES
// OR CONDITIONS OF ANY KIND, either express or implied.

package md5

import (
	"crypto/md5"
	"fmt"
	"io"
	"os"
	"wrs/tk/packages/checksum"

	"github.com/pkg/errors"
)

func Md5OfFilePath(filePath string) (string, error) {
	if filePath == "" {
		return EmptyMd5(), &checksum.EmptyChecksum{Algorithm: "md5"}
	}

	f, err := os.Open(filePath)
	if err != nil {
		return "", errors.Wrapf(err, "error calculating MD5 of %s", filePath)
	}
	defer f.Close()

	hasher := md5.New()
	if _, err := io.Copy(hasher, f); err != nil {
		return "", errors.Wrapf(err, "error writing %s to MD5 hasher", filePath)
	}
	hash := fmt.Sprintf("%x", hasher.Sum(nil))

	return hash, nil
}

func EmptyMd5() string {
	hasher := md5.New()
	hash := fmt.Sprintf("%x", hasher.Sum(nil))

	return hash
}
