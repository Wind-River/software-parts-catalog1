// Copyright (c) 2020 Wind River Systems, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at:
//       http://www.apache.org/licenses/LICENSE-2.0
// Unless required by applicable law or agreed to in writing, software  distributed
// under the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES
// OR CONDITIONS OF ANY KIND, either express or implied.

package license_web

import (
	"encoding/hex"
	"encoding/json"
	"net/http"
	"strings"
	"time"
	"wrs/tk/packages/core/archive"
	"wrs/tk/packages/core/file_collection"
	"wrs/tk/packages/core/license"

	"github.com/rs/zerolog/log"
)

type License struct {
	Count     int `json:"count"`
	Container struct {
		ContainerID int64  `json:"container_id"`
		Name        string `json:"container_name"`
		GroupID     int64  `json:"group_container_id,omitempty"`
		Group       string `json:"group_name,omitempty"`
	} `json:"container"`
	Licenses    []license.License `json:"licenses"`
	RequestTime time.Time         `json:"request_time"`
}

func return404(w http.ResponseWriter, message string) {
	var ret License

	ret.Count = 0
	ret.Licenses = []license.License{}

	if err := json.NewEncoder(w).Encode(ret); err != nil {
		http.Error(w, "error encoding json response", 500)
		return
	}

	// http.Error(w, message, 404)
}

func HandleLicenseQuery(w http.ResponseWriter, r *http.Request) {
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

	licenseController, err := license.GetLicenseController(r)
	if err != nil {
		http.Error(w, "error getting license controller", 500)
		return
	}

	var ret License
	ret.RequestTime = now

	fileCollection, err := fileCollectionController.GetBy(vcode, 0)
	if err != nil && err != file_collection.ErrNotFound {
		log.Error().Err(err).Str("vcode", hex.EncodeToString(vcode)).Msg("error getting file_collection")
		http.Error(w, "error getting file_collection", 500)
		return
	}
	arch, err := archiveController.GetBy(shaTwo, shaOne, name)
	if err != nil && err != archive.ErrNotFound {
		http.Error(w, "error getting archive", 500)
		return
	} else if err != nil { // make sure archiv.ErrNotFound is ignored, and not caught by a later check
		err = nil
	}

	if fileCollection == nil && arch == nil {
		return404(w, "container not found")
		return
	} else if fileCollection == nil && arch != nil {
		if !arch.FileCollectionID.Valid {
			return404(w, "file_collection not found")
			return
		}

		fileCollection, err = fileCollectionController.GetByID(arch.FileCollectionID.Int64)
		if err != nil {
			fileCollection, err = fileCollectionController.GetByID(arch.FileCollectionID.Int64)
			if err == file_collection.ErrNotFound {
				return404(w, "file_collection not found")
				return
			} else if err != nil {
				http.Error(w, "error getting file_collection", 500)
				return
			}
		}
	} else if arch != nil && arch.FileCollectionID.Valid && arch.FileCollectionID.Int64 != fileCollection.FileCollectionID {
		// ignore archive that does not match file collection
		arch = nil
	}

	ret.Container.ContainerID = fileCollection.FileCollectionID
	if fileCollection.GroupID.Valid {
		ret.Container.GroupID = fileCollection.GroupID.Int64
		ret.Container.Group = fileCollection.GroupName.String
	}
	if arch != nil {
		ret.Container.Name = arch.Name.String
	}

	var fcLicense *license.License
	if fileCollection.LicenseID.Valid {
		log.Debug().Int64("license_id", fileCollection.LicenseID.Int64).Msg("GetByID")
		fcLicense, err = licenseController.GetByID(fileCollection.LicenseID.Int64)
	} else if fileCollection.GroupID.Valid {
		log.Debug().Int64("group_id", fileCollection.GroupID.Int64).Msg("GetByGroup")
		fcLicense, err = licenseController.GetByGroup(fileCollection.GroupID.Int64)
	}
	if err != nil {
		log.Error().Err(err).Int64("file_collection_id", fileCollection.FileCollectionID).Msg("error getting license")
		http.Error(w, "error getting license", 500)
		return
	}

	if fcLicense == nil {
		ret.Count = 0
		ret.Licenses = []license.License{}
	} else {
		ret.Count = 1
		ret.Licenses = []license.License{*fcLicense}
	}

	if err := json.NewEncoder(w).Encode(ret); err != nil {
		http.Error(w, "error encoding json response", 500)
		return
	}
}
