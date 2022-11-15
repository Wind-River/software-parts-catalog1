// Copyright (c) 2020 Wind River Systems, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at:
//       http://www.apache.org/licenses/LICENSE-2.0
// Unless required by applicable law or agreed to in writing, software  distributed
// under the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES
// OR CONDITIONS OF ANY KIND, either express or implied.

package unicode

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestToValidUTF8(t *testing.T) {
	src, err := filepath.Abs("./test/src/files")
	if err != nil {
		t.Skipf("Could not find source directory: %s", err.Error())
	}

	files, err := os.ReadDir(src)
	if err != nil {
		t.Skipf("Could not read source directory <%s>: %s", src, err.Error())
	}

	for _, v := range files {
		_ = ToValidUTF8(v.Name())
	}

	src, err = filepath.Abs("./test/src/data/invalid.test.txt")
	if err != nil {
		t.Skipf("Could not find source file invalid.test.txt: %s", err.Error())
	}

	bytes, err := os.ReadFile(src)
	if err != nil {
		t.Skipf("Could not read data file invalid.test.txt: %s", err.Error())
	}

	lines := strings.Split(string(bytes), "\n")
	if len(lines) == 0 {
		t.Skipf("Data file invalid.test.txt contains no lines")
	}

	for _, v := range lines {
		_ = ToValidUTF8(v)
	}
}

func TestNotActuallyATest(t *testing.T) {

	output := ""
	value := []byte("ΓÖ¬ΓÖ¼")
	for _, v := range value {
		output = fmt.Sprintf("%s\\x%x", output, v)
	}
	t.Logf("%s\n", output)

	// To make sure logs are printed
	t.Fail()
}
