// Copyright (c) 2020 Wind River Systems, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at:
//       http://www.apache.org/licenses/LICENSE-2.0
// Unless required by applicable law or agreed to in writing, software  distributed
// under the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES
// OR CONDITIONS OF ANY KIND, either express or implied.

package file_collection

import (
	"context"

	"github.com/pkg/errors"
)

type Key int

// FileCollectionKey guarentees uniqueness for use as a context value key.
const FileCollectionKey Key = iota

// Get FileCollectionController or return an error
func GetFileCollectionController(ctx context.Context) (*FileCollectionController, error) {
	switch contextValue := ctx.Value(FileCollectionKey).(type) {
	case *FileCollectionController:
		if contextValue == nil {
			return nil, errors.New("FileCollectionController is nil")
		}

		return contextValue, nil
	case nil: // not found
		return nil, errors.New("FileCollectionController not found")
	default:
		return nil, errors.Wrapf(errors.New("unexpected type"), "got %#v", contextValue)
	}
}
