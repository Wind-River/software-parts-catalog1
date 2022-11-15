// Copyright (c) 2020 Wind River Systems, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at:
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software  distributed
// under the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES
// OR CONDITIONS OF ANY KIND, either express or implied.

package aes

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"os"

	"github.com/pkg/errors"
)

// Generate, encode, and decode AES keys

type Key []byte

func GenerateKey() Key {
	token := make([]byte, 32)
	rand.Read(token)

	return token
}

func KeyFromBytes(in []byte) Key {
	return in
}

func KeyFromHexString(in string) (Key, error) {
	b, err := hex.DecodeString(in)
	if err != nil {
		return nil, errors.Wrap(err, "error decoding hex")
	}

	return KeyFromBytes(b), nil
}

func KeyFromBase64String(in string) (Key, error) {
	b, err := base64.URLEncoding.DecodeString(in)
	if err != nil {
		return nil, errors.Wrap(err, "error decoding base64")
	}

	return KeyFromBytes(b), nil
}

func KeyFromFile(filePath string, base int) (Key, error) {
	rawBytes, err := os.ReadFile(filePath)
	if err != nil {
		return nil, errors.Wrap(err, "error reading key file")
	}

	switch base {
	case 0:
		return KeyFromBytes(rawBytes), nil
	case 16:
		return KeyFromHexString(string(rawBytes))
	case 64:
		return KeyFromBase64String(string(rawBytes))
	default:
		return nil, errors.New(fmt.Sprintf("Unsupported base: %d", base))
	}
}

func (k Key) Encode(base int) []byte {
	switch base {
	case 0:
		return k
	case 16:
		ret := make([]byte, hex.EncodedLen(len(k)))
		hex.Encode(ret, k)
		return ret
	case 64:
		ret := make([]byte, base64.URLEncoding.EncodedLen(len(k)))
		base64.URLEncoding.Encode(ret, k)
		return ret
	default:
		return k
	}
}
