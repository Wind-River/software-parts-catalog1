// Copyright (c) 2020 Wind River Systems, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at:
//       http://www.apache.org/licenses/LICENSE-2.0
// Unless required by applicable law or agreed to in writing, software  distributed
// under the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES
// OR CONDITIONS OF ANY KIND, either express or implied.

package config

// MainConfig defines the values expected from a toml configuration file.
// All values are publically available so that they can be overridden by command line options.
type MainConfig struct {
	Server struct {
		Port            int    `toml:"port"`   // Port for the server to bind to
		UploadDirectory string `toml:"upload"` // Directory to initially store archives uploaded to TK
		DaemonMode      bool   `toml:"daemon"` // Daemon mode resolves relative paths using the binary's directory

		Threads          int    `toml:"threads"` // Max number of threads to use
		SecretsDirectory string `toml:"secrets"` // Directory to find sensitive secrets
	} `toml:"server"`

	Blob struct { // Configuration for object-storage
		Endpoint string `toml:"endpoint"`
		Region   string `toml:"region"`
		Bucket   string `toml:"bucket"`
		ID       string `toml:"id"`
		Secret   string `toml:"secret"`
		Token    string `toml:"token"`
	} `toml:"blob"`

	Frontdoor struct { // Frontdoor host to download source from
		Host string `toml:"host"`
	}

	Search struct {
		InsertCost     int `toml:"insert"`
		Deletecost     int `toml:"delete"`
		SubstituteCost int `toml:"substitute"`
		MaxDistance    int `toml:"maxDistance"`
	} `toml:"search"`
}

// Initialize a configuration with defaults set
func DefaultConfig() *MainConfig {
	ret := new(MainConfig)
	ret.Server.Port = 4200
	ret.Server.Threads = 1

	return ret
}
