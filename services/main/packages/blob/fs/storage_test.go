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
	"crypto/sha1"
	"crypto/sha256"
	"io"
	"os"
	"testing"
	"wrs/tk/packages/blob/file"
)

func TestBlobFileSystem_Store_Retrive(t *testing.T) {
	tmp, err := os.MkdirTemp("", "TestBlobFileSystem")
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	defer os.RemoveAll(tmp)

	fs, err := CreateBlobFileSystem(tmp)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	type fields struct {
		fs *BlobFileSystem
	}
	type args struct {
		data     []byte
		mimetype string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			"date file",
			fields{fs},
			args{
				[]byte("Wed 14 Apr 2021 05:44:08 PM UTC"),
				"plain/text",
			},
			false,
		},
		{
			"foo bar zap",
			fields{fs},
			args{
				[]byte("foo\nbar\nzap\n"),
				"plain/text",
			},
			false,
		},
		{
			"date file again",
			fields{fs},
			args{
				[]byte("Wed 14 Apr 2021 05:44:08 PM UTC"),
				"plain/text",
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			size := len(tt.args.data)
			var sha1 file.Sha1 = sha1.Sum(tt.args.data)
			var sha256 file.Sha256 = sha256.Sum256(tt.args.data)

			if err := fs.Store(bytes.NewReader(tt.args.data), &file.FileInfo{
				Size:     int64(size),
				MimeType: tt.args.mimetype,
				Sha256:   sha256,
				Sha1:     sha1,
			}); (err != nil) != tt.wantErr {
				t.Errorf("BlobFileSystem.Store() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			f, err := fs.Retrieve(sha256)
			if (err != nil) != tt.wantErr {
				t.Errorf("BlobFileSystem.Retrieve() error = %+v, wantErr %v", err, tt.wantErr)
				return
			}

			got, err := io.ReadAll(f)
			if err != nil {
				t.Error(err)
				return
			}

			if !bytes.Equal(tt.args.data, got) {
				t.Errorf("BlobFileSystem.Retrieve() = %v, want %v", got, tt.args.data)
				return
			}
		})
	}
}
