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
	"unicode/utf8"
)

func ToValidUTF8(s string) string {
	if utf8.ValidString(s) {
		return s
	}

	b := []byte(s)
	ret := make([]rune, 0, len(s)*3)
	for i, v := range b {
		if r, _ := utf8.DecodeRune(b[i:]); r != utf8.RuneError {
			ret = append(ret, r)
		} else {
			ret = append(ret, []rune{'\\', 'x'}...)
			ret = append(ret, []rune(fmt.Sprintf("%x", v))...)
		}
	}

	return string(ret)
}
