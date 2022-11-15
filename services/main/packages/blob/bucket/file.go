// Copyright (c) 2020 Wind River Systems, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at:
//       http://www.apache.org/licenses/LICENSE-2.0
// Unless required by applicable law or agreed to in writing, software  distributed
// under the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES
// OR CONDITIONS OF ANY KIND, either express or implied.

package bucket

import (
	"os"

	"github.com/pkg/errors"
)

// TransientFile is meant to be a purely temporary file that gets cleaned up on Close
type TransientFile struct {
	*os.File
}

func (f *TransientFile) Close() error {
	name := f.Name()

	if err := f.File.Close(); err != nil {
		return errors.Wrapf(err, "error closing temporary file")
	}

	if err := os.Remove(name); err != nil {
		return errors.Wrapf(err, "error removing temporary file")
	}

	return nil
}
