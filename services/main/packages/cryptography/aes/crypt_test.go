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
	"os/exec"
	"testing"

	"github.com/pkg/errors"
)

// TestEncryptDecrypt tests golang Encrypt and Decrypt text
func TestEncryptDecrypt(t *testing.T) {
	key := GenerateKey()
	plaintext := "The quick brown fox jumps over the lazy dog"

	encryptedData, err := Encrypt(key, []byte(plaintext))
	if err != nil {
		t.Logf("%+v\n", err)
		t.FailNow()
	}

	decryptedData, err := Decrypt(key, encryptedData, 0)
	if err != nil {
		t.Logf("%+v\n", err)
		t.FailNow()
	}

	if plaintext != string(decryptedData) {
		t.Logf("%s != %s\n", plaintext, string(decryptedData))
		t.FailNow()
	}
}

// TestPHPEncryptDecrypt tests PHP Encrypt and Decrypt text
func TestPHPEncryptDecrypt(t *testing.T) {
	key := GenerateKey()
	plaintext := "The quick brown fox jumps over the lazy dog"

	encrypted, err := exec.Command("php", "../../test/php/encrypt.php", string(key.Encode(64)), plaintext).Output()
	if err != nil {
		t.Logf("%+v", errors.Wrap(err, "error PHP encrypting"))
		t.FailNow()
	}

	decrypted, err := exec.Command("php", "../../test/php/decrypt.php", string(key.Encode(64)), string(encrypted)).Output()
	if err != nil {
		t.Logf("%+v", errors.Wrap(err, "error PHP decrypting"))
		t.FailNow()
	}

	if plaintext != string(decrypted) {
		t.Logf("%s != %s\n", plaintext, string(decrypted))
		t.FailNow()
	}
}

// TestPHPCrossCrypt tests encrypting in PHP and decrypting in golang, and vice versa
func TestPHPCrossCrypt(t *testing.T) {
	key := GenerateKey()
	plainA := "The quick brown fox jumps over the lazy dog"
	plainB := "Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum."

	encryptedA, err := exec.Command("php", "../../test/php/encrypt.php", string(key.Encode(64)), plainA).Output()
	if err != nil {
		t.Logf("%+v", errors.Wrap(err, "error PHP encrypting plainA"))
		t.FailNow()
	}

	decryptedA, err := Decrypt(key, AESDataFromHexBytes(encryptedA), 0)
	if err != nil {
		t.Logf("%+v", errors.Wrap(err, "error golang decrypting encryptedA"))
		t.FailNow()
	}

	if plainA != string(decryptedA) {
		t.Logf("%s =PHP=> %x =GOLANG=> %s\n", plainA, encryptedA, string(decryptedA))
		t.FailNow()
	}

	encryptedB, err := Encrypt(key, []byte(plainB))
	if err != nil {
		t.Logf("%+v", errors.Wrap(err, "error golang encrypting plainB"))
		t.FailNow()
	}

	decryptedB, err := exec.Command("php", "../../test/php/decrypt.php", string(key.Encode(64)), encryptedB.String()).Output()
	if err != nil {
		t.Logf("%+v", errors.Wrap(err, "error PHP decrypted encryptedB"))
		t.FailNow()
	}

	if plainB != string(decryptedB) {
		t.Logf("%s =GOLANG=> %x =PHP=> %s\n", plainB, encryptedB, string(decryptedB))
		t.FailNow()
	}
}
