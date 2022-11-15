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
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"

	"github.com/pkg/errors"
)

// The actual decryption and encryption functions

func Decrypt(key Key, data *AESEncryptedData, base int) ([]byte, error) {
	// log.Printf("Decrypt(%x, %x, %d)", key, data.Bytes(), base)
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, errors.Wrap(err, "error creating cipher")
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, errors.Wrap(err, "error creating counter")
	}

	// crypto/aes expects payload+tag
	ciphertext := make([]byte, 0, len(data.payload)+len(data.tag))
	ciphertext = append(ciphertext, data.payload...)
	ciphertext = append(ciphertext, data.tag...)

	ret, err := aesgcm.Open(nil, data.nonce, ciphertext, nil)
	if err != nil {
		return ret, errors.Wrap(err, "error decrypting")
	}

	switch base {
	case 0:
		return ret, nil
	case 16:
		h := make([]byte, hex.EncodedLen(len(ret)))
		hex.Encode(h, ret)

		return h, nil
	case 64:
		b := make([]byte, base64.URLEncoding.EncodedLen(len(ret)))
		base64.URLEncoding.Encode(b, ret)

		return b, nil
	default:
		return nil, errors.New(fmt.Sprintf("Unexpected base: %d", base))
	}
}

func Encrypt(key Key, payload []byte) (*AESEncryptedData, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, errors.Wrap(err, "error creating cipher")
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, errors.Wrap(err, "error creating counter")
	}

	nonce := make([]byte, 12)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, errors.Wrap(err, "error creating nonce")
	}

	ciphertext := aesgcm.Seal(nil, nonce, payload, nil)

	ret := new(AESEncryptedData)
	ret.nonce = nonce
	ret.tag = make([]byte, 16)
	copy(ret.tag, ciphertext[len(ciphertext)-16:])
	ret.payload = make([]byte, len(ciphertext)-16)
	copy(ret.payload, ciphertext[:len(ciphertext)-16])

	return ret, nil
}
