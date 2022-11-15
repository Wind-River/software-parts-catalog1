// Copyright (c) 2020 Wind River Systems, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at:
//       http://www.apache.org/licenses/LICENSE-2.0
// Unless required by applicable law or agreed to in writing, software  distributed
// under the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES
// OR CONDITIONS OF ANY KIND, either express or implied.

package upload_web

import (
	"mime/multipart"
	"os"
	"path/filepath"
)

// The data that gets sent to the frontend as responses

// UploadPayload is the object the frontend expects after an upload
type UploadPayload struct {
	Filename    string                `json:"Filename"`
	Uploadname  string                `json:"Uploadname"`
	Sha1        string                `json:"Sha1,omitempty"`
	ContentType string                `json:"Content-Type"`
	IsMeta      bool                  `json:"isMeta"`
	RawHeader   *multipart.FileHeader `json:"Header"`
	Extra       string                `json:"Extra,omitempty"`
}

// ToPayload implements Payloader
func (u UploadPayload) ToPayload() *UploadPayload {
	return &u
}

// Payloader allows different upload handlers to return the same response structure to the frontend.
// Currently only expect CSV for data loading, or archive for extraction and processing
type Payloader interface {
	ToPayload() *UploadPayload
}

// HandleFile is the payload returned from processing an archive
type HandledFile struct {
	File   *os.File
	Sha1   string
	Header *multipart.FileHeader
}

func (h HandledFile) ToPayload() *UploadPayload {
	return &UploadPayload{
		h.Header.Filename,
		filepath.Base(h.File.Name()),
		h.Sha1,
		h.Header.Header.Get("Content-Type"),
		false,
		h.Header,
		"",
	}
}

// processRequestFile is the json request frontend sends after it is done uploading archives.
type processRequestFile struct {
	ID          string `json:"id"`
	Filename    string `json:"filename"`
	Uploadname  string `json:"uploadname"`
	ContentType string `json:"contentType"`

	Filepath string
}
