// Copyright (c) 2020 Wind River Systems, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at:
//       http://www.apache.org/licenses/LICENSE-2.0
// Unless required by applicable law or agreed to in writing, software  distributed
// under the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES
// OR CONDITIONS OF ANY KIND, either express or implied.

package server

import (
	"context"
	"fmt"
	"net/http"
	"path/filepath"
	"wrs/tk/packages/blob"
	"wrs/tk/packages/blob/bucket"
	mainConfig "wrs/tk/packages/config"
	archive_core "wrs/tk/packages/core/archive"
	"wrs/tk/packages/core/part"
	"wrs/tk/packages/core/partlist"
	"wrs/tk/packages/web_services/archive_web"

	// "wrs/tk/packages/core/group"
	"wrs/tk/packages/core/license"
	"wrs/tk/packages/database"
	"wrs/tk/packages/middleware"

	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/go-chi/chi/v5"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"wrs/tk/packages/graphql"
	"wrs/tk/packages/graphql/generated"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
)

type Server struct {
	http.Server
	running bool
	logger  *zerolog.Logger
}

func NewServer(host string, port int, db *sqlx.DB, frontdoorHost string, threads int, config *mainConfig.MainConfig, logger *zerolog.Logger) (*Server, error) {
	// Transform input if necessary
	if threads == 0 {
		threads = 1
	}

	address := host
	if port > 0 {
		address = fmt.Sprintf("%s:%d", address, port)
	}

	// Set-up server with address
	server := Server{
		Server: http.Server{
			Addr: address,
		},
	}

	if logger == nil {
		server.logger = &log.Logger
	} else {
		server.logger = logger
	}

	if config.Blob.Bucket == "" {
		return nil, errors.New("no bucket found")
	}
	// Initialize requirements for Handler
	// initialize blob storage
	var fileStorage blob.Storage
	var archiveStorage blob.Storage
	log.Info().Str("config.Blob.Bucket", config.Blob.Bucket).Msg("setting up s3 blob storage")

	// connect to blob database
	secrets := "/var/run/secrets"
	info, err := database.NewEncryptedDBInfo(filepath.Join(secrets, "blob"), filepath.Join(secrets, "key"))
	if err != nil {
		return nil, errors.Wrapf(err, "error reading blob database info")
	}
	blobDB, err := info.Connect()
	if err != nil {
		return nil, errors.Wrapf(err, "error connecting to blob database")
	}

	// if credentials required create credentials object
	// a nil cred object can still be a valid use of credentials, as it would then rely on IAM roles
	var cred *credentials.Credentials
	if config.Blob.ID != "" {
		log.Info().Str("ObjectStorage.ID", config.Blob.ID).Msg("Using ObjectStorage Credentials")
		cred = credentials.NewStaticCredentials(config.Blob.ID, config.Blob.Secret, config.Blob.Token)
	}

	// create blob storage
	fileStorage, err = bucket.CreateBlobBucket(blobDB, "blob", config.Blob.Bucket, config.Blob.Region, config.Blob.Endpoint, cred)
	if err != nil {
		err = errors.Wrapf(err, "error opening blob bucket")
		return nil, err
	}
	archiveStorage, err = bucket.CreateBlobBucket(blobDB, "archive", config.Blob.Bucket, config.Blob.Region, config.Blob.Endpoint, cred)
	if err != nil {
		err = errors.Wrapf(err, "error opening blob bucket")
		return nil, err
	}

	// Create http.Handler
	router := chi.NewRouter()

	server.Server.Handler = router

	// Create new controllers
	archiveController := archive_core.NewArchiveController(db, fileStorage, archiveStorage, int(threads), config.Blob.Bucket, cred, config.Blob.Endpoint, config.Blob.Region)
	partController := part.PartController{DB: db}
	partlistController := partlist.PartListController{DB: db}
	licenseController := license.LicenseController{
		DB:                db,
		PartController:    partController,
		ArchiveController: archiveController,
	}
	// groupController := group.GroupController{DB: db}

	router.Use(middleware.ContextWithValue(archive_core.ArchiveKey, archiveController))
	router.Use(middleware.ContextWithValue(part.PartKey, &partController))
	router.Use(middleware.ContextWithValue(partlist.PartListKey, &partlistController))
	router.Use(middleware.ContextWithValue(license.LicenseKey, &licenseController))
	// router.Use(middleware.ContextWithValue(group.GroupKey, &groupController))

	//
	graphqlHandler := handler.NewDefaultServer(generated.NewExecutableSchema(generated.Config{Resolvers: &graphql.Resolver{
		ArchiveController:  archiveController,
		PartController:     &partController,
		PartListController: &partlistController,
		LicenseController:  &licenseController,
	}}))
	router.Handle("/playground", playground.Handler("GraphQL playground", "/api/graphql"))
	router.Handle("/api/graphql", graphqlHandler)
	router.Get("/api/archive/{archiveSha256:[a-fA-F0-9]+}", archive_web.HandleArchiveDownload)               // if archive has a name, which it probably does, redirects
	router.Get("/api/archive/{archiveSha256:[a-fA-F0-9]+}/{archiveName}", archive_web.HandleArchiveDownload) // serves archive with the given name

	return &server, nil
}

// ListenAndServe should handle any start-up not already part of the router, and then pass execution to the http.Server
func (server *Server) ListenAndServe() error {
	if server.running {
		server.logger.Warn().Interface("server", server).Msg("Server already running")
		return nil
	}

	return server.Server.ListenAndServe()
}

// Shutdown should handle any shutdown not already parto the router, and the pass execution to the http.Server
func (server *Server) Shutdown(ctx context.Context) error {
	if !server.running {
		server.logger.Warn().Interface("server", server).Msg("Server already shutdown")
		return nil
	}

	if err := server.Server.Shutdown(ctx); err != nil {
		return err
	}

	server.running = false
	return nil
}
