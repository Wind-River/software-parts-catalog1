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
	"testing"
	"wrs/tk/packages/cryptography/aes"

	_ "github.com/jackc/pgx/v4/stdlib"
)

func TestDBInfo_Encrypt(t *testing.T) {
	type fields struct {
		DBName      string
		User        string
		Password    string
		passwordRaw string
		Host        string
		Port        int
	}
	type args struct {
		key aes.Key
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "Generate encrypted info",
			fields: fields{
				DBName:      "tkdb",
				User:        "postgres",
				passwordRaw: "sql",
				Host:        "localhost",
				Port:        5432,
			},
			args:    args{key: aes.GenerateKey()},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &DBInfo{
				DBName:      tt.fields.DBName,
				User:        tt.fields.User,
				Password:    tt.fields.Password,
				passwordRaw: tt.fields.passwordRaw,
				Host:        tt.fields.Host,
				Port:        tt.fields.Port,
			}
			if err := d.Encrypt(tt.args.key); (err != nil) != tt.wantErr {
				t.Errorf("DBInfo.Encrypt() error = %v, wantErr %v", err, tt.wantErr)
				t.Errorf(d.Password)
			}
		})
	}
}
