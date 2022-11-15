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
	"testing"
)

func TestSha256_IsValid(t *testing.T) {
	tests := []struct {
		name string
		h    Sha256
		want bool
	}{
		{
			name: "NULL",
			h:    Sha256{},
			want: false,
		},
		{
			name: "EMPTY",
			h:    Sha256{227, 176, 196, 66, 152, 252, 28, 20, 154, 251, 244, 200, 153, 111, 185, 36, 39, 174, 65, 228, 100, 155, 147, 76, 164, 149, 153, 27, 120, 82, 184, 85},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.h.IsValid(); got != tt.want {
				t.Errorf("Sha256.IsNull() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSha256_IsEmpty(t *testing.T) {
	tests := []struct {
		name string
		h    Sha256
		want bool
	}{
		{
			name: "NULL",
			h:    Sha256{},
			want: false,
		},
		{
			name: "EMPTY",
			h:    Sha256{227, 176, 196, 66, 152, 252, 28, 20, 154, 251, 244, 200, 153, 111, 185, 36, 39, 174, 65, 228, 100, 155, 147, 76, 164, 149, 153, 27, 120, 82, 184, 85},
			want: true,
		}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.h.IsEmpty(); got != tt.want {
				t.Errorf("Sha256.IsEmpty() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSha1_IsValid(t *testing.T) {
	tests := []struct {
		name string
		h    Sha1
		want bool
	}{
		{
			name: "NULL",
			h:    Sha1{},
			want: false,
		},
		{
			name: "EMPTY",
			h:    Sha1{218, 57, 163, 238, 94, 107, 75, 13, 50, 85, 191, 239, 149, 96, 24, 144, 175, 216, 7, 9},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.h.IsValid(); got != tt.want {
				t.Errorf("Sha1.IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSha1_IsEmpty(t *testing.T) {
	tests := []struct {
		name string
		h    Sha1
		want bool
	}{
		{
			name: "NULL",
			h:    Sha1{},
			want: false,
		},
		{
			name: "EMPTY",
			h:    Sha1{218, 57, 163, 238, 94, 107, 75, 13, 50, 85, 191, 239, 149, 96, 24, 144, 175, 216, 7, 9},
			want: true,
		}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.h.IsEmpty(); got != tt.want {
				t.Errorf("Sha1.IsEmpty() = %v, want %v", got, tt.want)
			}
		})
	}
}
