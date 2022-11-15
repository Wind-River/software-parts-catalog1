// Copyright (c) 2020 Wind River Systems, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at:
//       http://www.apache.org/licenses/LICENSE-2.0
// Unless required by applicable law or agreed to in writing, software  distributed
// under the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES
// OR CONDITIONS OF ANY KIND, either express or implied.

package database

import (
	"encoding/json"
	"fmt"
	"os"
	"wrs/tk/packages/cryptography/aes"

	_ "github.com/jackc/pgx/v4/stdlib"

	"github.com/jmoiron/sqlx"
)

// DBInfo contains all the fields necessary to construct a connection string
type DBInfo struct {
	DBName      string `json:"dbname"`
	User        string `json:"user"`
	Password    string `json:"password"`
	passwordRaw string
	Host        string `json:"host"`
	Port        int    `json:"port"`
}

// Decrypt using the given aes key to decrypt a database passward
// passwordRaw is decrypted into Password
func (d *DBInfo) Decrypt(key aes.Key) error {
	d.passwordRaw = d.Password
	if len(d.Password) < 24 {
		return fmt.Errorf("encrypted password too short: %s", d.Password)
	}

	ciphertext := aes.AESDataFromHexBytes([]byte(d.Password))

	plainbytes, err := aes.Decrypt(key, ciphertext, 0)
	if err != nil {
		return err
	}

	d.Password = string(plainbytes)

	return nil
}

// Encrypt uses the given aes key to encrypt a database password
// passwordRaw is encrypted into Password
func (d *DBInfo) Encrypt(key aes.Key) error {
	d.passwordRaw = d.Password

	payload, err := aes.Encrypt(key, []byte(d.passwordRaw))
	if err != nil {
		return err
	}

	d.Password, err = payload.Format(16)
	if err != nil {
		return nil
	}

	return nil
}

// Connect creates the connection string and connects to the database
func (d *DBInfo) Connect() (*sqlx.DB, error) {
	connStr := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
		d.User,
		d.Password,
		d.Host,
		d.Port,
		d.DBName,
	)

	db, err := sqlx.Open("pgx", connStr)

	return db, err
}

// NewDBInfo loads a DBInfo from a file found at DBInfoPath
func NewDBInfo(DBInfoPath string) (*DBInfo, error) {
	body, err := os.ReadFile(DBInfoPath)
	if err != nil {
		return nil, err
	}

	var d DBInfo
	err = json.Unmarshal(body, &d)

	return &d, err
}

// NewEncryptedDBInfo loads a DBInfo from a file and then decrypts its password using an aes key from a file
func NewEncryptedDBInfo(dbInfoPath string, keyPath string) (*DBInfo, error) {
	info, err := NewDBInfo(dbInfoPath)
	if err != nil {
		return info, err
	}

	key, err := aes.KeyFromFile(keyPath, 64)
	if err != nil {
		return info, err
	}

	err = info.Decrypt(key)
	return info, err
}
