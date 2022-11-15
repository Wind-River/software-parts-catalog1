//go:build !script

// Copyright (c) 2020 Wind River Systems, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at:
//       http://www.apache.org/licenses/LICENSE-2.0
// Unless required by applicable law or agreed to in writing, software  distributed
// under the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES
// OR CONDITIONS OF ANY KIND, either express or implied.

// Package main parses cli options and config files before passing it off to the server package.
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	mainConfig "wrs/tk/packages/config"

	"wrs/tk/packages/database"
	"wrs/tk/packages/server"

	"github.com/BurntSushi/toml"
	"github.com/rs/zerolog/log"
)

// if versionFlag is set, print version and exit
var versionFlag bool

// if helpFlag is set, print usage and exit
var helpFlag bool

// port for server to bind to
var argPort int

// directory to store and process uploaded files
var argUploadDirectory string

// number of threads that should be used
var argThreads int

// host for frontdoor to ask for source
// should be removed once agent in place
var argFrontdoorHost string

// path to configuration file
var argConfigPath string

// path to sensitive directory
var argSecretsDirectory string

// configuration struct
var config *mainConfig.MainConfig

func init() {
	flag.IntVar(&argPort, "p", 80, "Port")
	flag.StringVar(&argUploadDirectory, "u", "", "Upload Directory")

	flag.BoolVar(&versionFlag, "version", false, "Print golang runtime version")
	flag.BoolVar(&helpFlag, "help", false, "Print defaults")
	flag.IntVar(&argThreads, "threads", 0, "Processor Max Threads")

	flag.StringVar(&argFrontdoorHost, "frontdoor", "", "Frontdoor Host")

	flag.StringVar(&argConfigPath, "config", "/highlander/config/config.toml", "Config Path")

	flag.StringVar(&argSecretsDirectory, "secrets", "", "Secrets Directory")

	// default configuration that will be overridden by configuration file and/or command line options
	config = mainConfig.DefaultConfig()
}

// setConfig loads the config file defined by argConfigPath, then deals with collisions between config file settings and command line options.
// Command line options always override config file options.
func setConfig() {
	// Load config file if exists
	if argConfigPath != "" {
		if metadata, err := toml.DecodeFile(argConfigPath, config); err != nil {
			log.Fatal().Str("argConfigPath", argConfigPath).Err(err).Send()
		} else {
			log.Info().Str("argConfigPath", argConfigPath).Interface("metadata", metadata).Interface("config", config).Msg("Loaded config file")
		}
	}

	// Overwrite config with command line arguments if set
	if argPort > 0 {
		config.Server.Port = argPort
	}
	if argUploadDirectory != "" {
		config.Server.UploadDirectory = argUploadDirectory
	}
	if argThreads > 0 {
		config.Server.Threads = argThreads
	}
	if argFrontdoorHost != "" {
		config.Frontdoor.Host = argFrontdoorHost
	}
	if argSecretsDirectory != "" {
		config.Server.SecretsDirectory = argSecretsDirectory
	}
}

// logFlags logs each options with final config value and the command line option.
func logFlags() {
	log.Info().
		Int("p", argPort).
		Int("Port", config.Server.Port).
		Msg("port")
	log.Info().
		Str("u", argUploadDirectory).
		Str("UploadDirectory", config.Server.UploadDirectory).
		Msg("argUploadDirectory")
	log.Info().
		Str("object.id", config.Blob.ID).
		Str("object.secret", config.Blob.Secret).
		Str("object.token", config.Blob.Token).
		Msg("object")
	log.Info().Bool("version", versionFlag).Msg("versionFlag")
	log.Info().Bool("help", helpFlag).Msg("helpFlag")
	log.Info().
		Int("threads", argThreads).
		Int("Threads", config.Server.Threads).
		Msg("argThreads")
	log.Info().
		Str("frontdoor", argFrontdoorHost).
		Str("FrontdoorHost", config.Frontdoor.Host).
		Msg("argFrontdoorHost")
}

// main parses command line options and formats arguments as necessary, before handing execution off the server
func main() {
	flag.Parse()  // parse cli options
	if helpFlag { // if cli is asking for help, print flag options and exit
		flag.PrintDefaults()
		os.Exit(0)
	}
	if versionFlag { // if cli is asking for version, print version string and exit
		fmt.Printf("golang runtime: %s\n", runtime.Version())
		os.Exit(0)
	}

	// merge config and options
	setConfig()
	logFlags()

	// verify valid blob settings
	if config.Blob.Bucket == "" {
		fmt.Printf("blob bucket missing")
		os.Exit(0)
	}

	// transform relative paths into absolute paths using the executable binary's location
	bin, err := os.Executable()
	if err != nil {
		log.Fatal().Err(err).Send()
	}

	dir := filepath.Dir(bin)
	if !filepath.IsAbs(config.Server.UploadDirectory) {
		_, err := os.Stat(config.Server.UploadDirectory)
		if err != nil {
			config.Server.UploadDirectory = filepath.Join(dir, config.Server.UploadDirectory)
		}
	}

	// load server database
	info, err := database.NewEncryptedDBInfo(filepath.Join(config.Server.SecretsDirectory, "db"),
		filepath.Join(config.Server.SecretsDirectory, "key"))
	if err != nil {
		log.Fatal().Err(err).Msg("error initializing database info")
	}

	db, err := info.Connect()
	if err != nil {
		log.Fatal().Err(err).Msg("error connecting to database")
	}

	// initialize server
	srv, err := server.NewServer("", argPort, db, config.Frontdoor.Host, config.Server.Threads, config, nil)
	if err != nil {
		log.Fatal().Err(err).Msg("error initializing server")
	}

	// run server
	if err := srv.ListenAndServe(); err != nil {
		log.Fatal().Err(err).Msg("server encountered error")
	}
}
