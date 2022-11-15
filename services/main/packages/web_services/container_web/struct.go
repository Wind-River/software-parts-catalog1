// Copyright (c) 2020 Wind River Systems, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at:
//       http://www.apache.org/licenses/LICENSE-2.0
// Unless required by applicable law or agreed to in writing, software  distributed
// under the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES
// OR CONDITIONS OF ANY KIND, either express or implied.

package container_web

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
	"wrs/tk/packages/core/archive"
	"wrs/tk/packages/core/file_collection"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"gopkg.in/guregu/null.v4"
)

type Container struct {
	RequestTime time.Time `json:"request_time"`

	GroupID   int64  `json:"group_container_id,omitempty" db:"group_container_id"`
	GroupName string `json:"group_name,omitempty" db:"group_name"`

	ArchiveID int64  `json:"aid" db:"aid"`
	Name      string `json:"name" db:"name"`
	Path      string `json:"path,omitempty" db:"size"`
	Size      int64  `json:"size,omitempty" db:"size"`
	Sha1      string `json:"checksum_sha1,omitempty" db:"checksum_sha1"`
	Sha256    string `json:"checksum_sha256,omitempty" db:"checksum_sha256"`
	Md5       string `json:"checksum_md5,omitempty" db:"checksum_md5"`

	ContainerID         int64  `json:"cid" db:"cid"`
	VerificationCodeOne []byte `json:"verification_code_one,omitempty" db:"verification_code_one"`
	VerificationCodeTwo []byte `json:"verification_code_two,omitempty" db:"verification_code_two"`
	LicenseID           int64  `json:"license_id,omitempty" db:"license_id"`
	LicenseRationale    string `json:"-" db:"license_rationale"`
	Extracted           bool   `json:"flag_extract" db:"flag_extract"`
	LicenseExtracted    bool   `json:"flag_license_extracted" db:"flag_license_extracted"`
}

// HandleContainerQuery handles searching for a container given an arhive name, archive checksum, or verification code
func HandleContainerQuery(w http.ResponseWriter, r *http.Request) {
	now := time.Now()
	queryValues := r.URL.Query()

	vcodes, hasVCode := queryValues["vcode"]
	var vcode []byte
	if hasVCode {
		vcode, _ = hex.DecodeString(vcodes[0])
	}

	hashes, hasHash := queryValues["hash"]
	var shaTwo, shaOne string
	if hasHash {
		for _, hash := range hashes {
			if strings.HasPrefix(hash, "SHA256:") {
				shaTwo = strings.TrimPrefix(hash, "SHA256:")
			} else if strings.HasPrefix(hash, "SHA1:") {
				shaOne = strings.TrimPrefix(hash, "SHA1:")
			} else if len(hash) == 40 && shaOne == "" { // don't overwrite an explicit sha1 with an implicit one
				shaOne = hash
			}
		}
	}
	names, hasName := queryValues["name"]
	var name string
	if hasName {
		name = names[0]
	}

	if !(hasHash || hasName) {
		http.Error(w, "No identifing information was found in the request", 400)
		return
	}

	fileCollectionController, err := file_collection.GetFileCollectionController(r.Context())
	if err != nil {
		http.Error(w, "error getting file_collection controller", 500)
		return
	}

	archiveController, err := archive.GetArchiveController(r.Context())
	if err != nil {
		http.Error(w, "error getting archive controller", 500)
		return
	}

	fileCollection, err := fileCollectionController.GetByVerificationCode(vcode)
	if err != nil && err != file_collection.ErrNotFound {
		http.Error(w, "error getting file_collection", 500)
		return
	}

	arch, err := archiveController.GetBy(shaTwo, shaOne, name)
	if err != nil && err != archive.ErrNotFound {
		http.Error(w, "error getting archive", 500)
		return
	}

	if fileCollection == nil && arch == nil {
		http.Error(w, "container not found", 404)
		return
	}

	if fileCollection == nil {
		// If no file collection found, but archive has a valid file collection id, load file collection
		if arch.FileCollectionID.Valid { // Get archive's file collection
			fileCollection, _ = fileCollectionController.GetByID(arch.FileCollectionID.Int64)
		}
	} else if arch != nil && arch.FileCollectionID.Int64 != fileCollection.FileCollectionID {
		// If an archive was found but does not match the found file collection, drop archive
		arch = nil
	}

	container := new(Container)
	container.RequestTime = now
	if fileCollection != nil {
		container.ContainerID = fileCollection.FileCollectionID
		if fileCollection.GroupID.Valid {
			container.GroupID = fileCollection.GroupID.Int64
		}
		container.GroupName = fileCollection.GroupName.String
		container.Extracted = fileCollection.Extracted
		container.LicenseExtracted = fileCollection.LicenseExtracted
		if fileCollection.LicenseID.Valid {
			container.LicenseID = fileCollection.LicenseID.Int64
		}
		container.LicenseRationale = fileCollection.LicenseRationale.String
		container.VerificationCodeOne = fileCollection.VerificationCodeOne
		container.VerificationCodeTwo = fileCollection.VerificationCodeTwo
	}
	if arch != nil {
		container.ArchiveID = arch.ArchiveID
		if arch.Name.Valid {
			container.Name = arch.Name.String
		}
		if arch.Path.Valid {
			container.Path = arch.Path.String
		}
		if arch.Size.Valid {
			container.Size = arch.Size.Int64
		}
		if arch.Sha256.Valid {
			container.Sha256 = arch.Sha256.String
		}
		if arch.Sha1.Valid {
			container.Sha1 = arch.Sha1.String
		}
		if arch.Md5.Valid {
			container.Md5 = arch.Md5.String
		}
		if !container.Extracted && arch.ExtractStatus > 0 {
			container.Extracted = true
		}
	}

	if err := json.NewEncoder(w).Encode(container); err != nil {
		http.Error(w, "error encoding response", 500)
		return
	}
}

type ShowArchive struct {
	ID   int64       `db:"id" json:"id"`
	Name string      `db:"name" json:"name"`
	Sha1 null.String `db:"checksum_sha1" json:"sha1"`
	Path null.String `db:"path" json:"path"`
}

type ShowContainer struct {
	Count      int64         `db:"count" json:"count"`
	InsertDate time.Time     `db:"insert_date" json:"date"`
	License    null.String   `db:"expression" json:"license"`
	Rationale  null.String   `db:"license_rationale" json:"rationale"`
	Archives   []ShowArchive `json:"archives"`
}

// HandleContainer returns information on a given file_collection
func HandleContainerGet(w http.ResponseWriter, r *http.Request) {
	fcidString := chi.URLParam(r, "fcid")
	fcid, err := strconv.ParseInt(fcidString, 10, 64)
	if err != nil {
		http.Error(w, fmt.Sprintf("could not parse id \"%s\"", fcidString), 400)
		return
	}

	fileCollectionController, err := file_collection.GetFileCollectionController(r.Context())
	if err != nil {
		http.Error(w, "error getting file_collection controller", 500)
		log.Error().Err(err).Msg("error getting file_collection controller")
		return
	}

	archiveController, err := archive.GetArchiveController(r.Context())
	if err != nil {
		http.Error(w, "error getting archive controller", 500)
		log.Error().Err(err).Msg("error getting archive controller")
		return
	}

	fileCollection, err := fileCollectionController.GetByID(fcid)
	if err != nil {
		if err == file_collection.ErrNotFound {
			http.Error(w, "file collection not found", 404)
			return
		}

		http.Error(w, "error getting file collection", 500)
		return
	}

	log.Info().Int64("file_collection_id", fcid).Msg("Getting File Collection")

	count, err := fileCollectionController.CountFiles(fileCollection.FileCollectionID)
	if err != nil {
		http.Error(w, "error counting files", 500)
		return
	}

	var ret ShowContainer
	ret.Count = count
	ret.InsertDate = fileCollection.InsertDate
	ret.License = null.NewString(fileCollection.LicenseExpression.String, fileCollection.LicenseExpression.Valid)
	ret.Rationale = null.NewString(fileCollection.LicenseRationale.String, fileCollection.LicenseRationale.Valid)
	ret.Archives = make([]ShowArchive, 0)

	archives, err := archiveController.GetByFileCollection(fileCollection.FileCollectionID)
	if err != nil {
		http.Error(w, "error selecting archives of file_collection", 500)
		return
	}

	for _, arch := range archives {
		ret.Archives = append(ret.Archives, ShowArchive{
			ID:   arch.ArchiveID,
			Name: arch.Name.String,
			Sha1: null.NewString(arch.Sha1.String, arch.Sha1.Valid),
			Path: null.NewString(arch.Path.String, arch.Path.Valid),
		})
	}

	if err := json.NewEncoder(w).Encode(ret); err != nil {
		http.Error(w, "error encoding JSON response", 500)
		return
	}
}

// HandleContainerDownload serves a given archive
func HandleContainerDownload(w http.ResponseWriter, r *http.Request) {
	// TODO download archive from bucket, and serve that
	log.Debug().Str(zerolog.CallerFieldName, "HandleContainerDownload").Send()
	aIDString := chi.URLParam(r, "archiveID")
	archiveID, err := strconv.ParseInt(aIDString, 10, 64)
	if err != nil {
		http.Error(w, fmt.Sprintf("could not parse id \"%s\"", aIDString), 400)
		return
	}

	archiveController, err := archive.GetArchiveController(r.Context())
	if err != nil {
		http.Error(w, "error getting archive controller", 500)
		log.Error().Err(err).Msg("error getting archive controller")
		return
	}

	arch, err := archiveController.GetByID(archiveID)
	if err == archive.ErrNotFound {
		log.Debug().Str(zerolog.CallerFieldName, "HandleContainerDownload").Int64("archive_id", archiveID).Msg("Returning 404 on missing archive")
		http.Error(w, "archive not found", 404)
		return
	} else if err != nil {
		http.Error(w, "error selecting archive by ID", 500)
		log.Error().Err(err).Msg("error selecting archive to serve")
		return
	}

	tmpDir, err := os.MkdirTemp("", "tk-serve")
	if err != nil {
		http.Error(w, "error serving file", 500)
		log.Error().Err(err).Msg("error serving file")
		return
	}
	defer os.RemoveAll(tmpDir)

	f, err := os.Create(filepath.Join(tmpDir, arch.Name.String))
	if err != nil {
		http.Error(w, "error serving file", 500)
		log.Error().Err(err).Msg("error creating file in temp directory")
		return
	}
	defer f.Close()

	if err := archiveController.DownloadTo(arch, f); err != nil {
		http.Error(w, "error downloading archive", 500)
		log.Error().Err(err).Msg("error downloading archive")
		return
	}
	if _, err := f.Seek(0, 0); err != nil {
		http.Error(w, "error serving archive", 500)
		log.Error().Err(err).Msg("error seeking archive")
		return

	}

	log.Debug().Str(zerolog.CallerFieldName, "HandleContainerDownload").Str("filepath", filepath.Join(tmpDir, arch.Name.String)).Msg("Serving File")
	http.ServeFile(w, r, filepath.Join(tmpDir, arch.Name.String)) // TODO catch 404
}
