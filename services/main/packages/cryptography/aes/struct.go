// Copyright (c) 2020 Wind River Systems, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at:
//       http://www.apache.org/licenses/LICENSE-2.0
// Unless required by applicable law or agreed to in writing, software  distributed
// under the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES
// OR CONDITIONS OF ANY KIND, either express or implied.

package aes

import (
	"encoding/base64"
	"encoding/hex"
	"fmt"

	"github.com/pkg/errors"
)

// TODO lookup best practices relating the the nonce
type AESEncryptedData struct {
	nonce   []byte // Number Used Once, or Initialization Vectory
	payload []byte // Encrypted Data
	tag     []byte // Authentication Tag, or Message Authentication Code
}

func AESDataFromBytes(in []byte) *AESEncryptedData {
	ret := new(AESEncryptedData)

	ret.nonce = make([]byte, 12)

	l := len(in) - 28
	if l < 0 {
		l = 0
	}
	ret.payload = make([]byte, l)

	ret.tag = make([]byte, 16)

	copy(ret.nonce, in[0:12])
	copy(ret.payload, in[12:len(in)-16])
	copy(ret.tag, in[len(in)-16:])

	return ret
}

func AESDataFromHexBytes(in []byte) *AESEncryptedData {
	mid := make([]byte, hex.DecodedLen(len(in)))
	if _, err := hex.Decode(mid, in); err != nil {
		return AESDataFromBytes(in)
	}

	return AESDataFromBytes(mid)
}

func NewAESData(nonce []byte, payload []byte, tag []byte) *AESEncryptedData {
	var ret AESEncryptedData

	ret.nonce = make([]byte, 0, len(nonce))
	copy(ret.nonce, nonce)

	ret.payload = make([]byte, 0, len(payload))
	copy(ret.payload, payload)

	ret.tag = make([]byte, 0, len(tag))
	copy(ret.tag, tag)

	return &ret
}

func (d AESEncryptedData) Nonce() []byte {
	return d.nonce
}

func (d AESEncryptedData) Payload() []byte {
	return d.payload
}

func (d AESEncryptedData) Tag() []byte {
	return d.tag
}

func (d AESEncryptedData) Bytes() []byte {
	ret := make([]byte, len(d.nonce))
	copy(ret, d.nonce)
	ret = append(ret, d.payload...)
	ret = append(ret, d.tag...)

	return ret
}

func (d AESEncryptedData) Format(base int) (string, error) {
	data := d.Bytes()

	switch base {
	case 0:
		return string(data), nil
	case 16:
		return hex.EncodeToString(data), nil
	case 64:
		return base64.URLEncoding.EncodeToString(data), nil
	default:
		return "", errors.New(fmt.Sprintf("Unexpected base: %d", base))
	}
}

func (d AESEncryptedData) String() string {
	ret, _ := d.Format(16)

	return ret
}
