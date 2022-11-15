// Copyright (c) 2020 Wind River Systems, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at:
//       http://www.apache.org/licenses/LICENSE-2.0
// Unless required by applicable law or agreed to in writing, software  distributed
// under the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES
// OR CONDITIONS OF ANY KIND, either express or implied.

package upload_web

import (
	"encoding/csv"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"path/filepath"

	"wrs/tk/packages/core/upload"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type UploadHandler struct {
	UploadController *upload.UploadController
}

func (handler UploadHandler) HandleUpload(w http.ResponseWriter, r *http.Request) {
	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Error Opening File", 400)
		log.Error().Err(err).Str(zerolog.CallerFieldName, "backend.upload()").Send()
		return
	}
	defer file.Close()

	var payload *UploadPayload

	// determine whether a csv or an archive was uploaded
	ext := filepath.Ext(header.Filename)
	// mimeType := header.Header.Get("Content-Type")

	switch ext {
	case ".csv": // CSV was uploaded
		payload, err = handler.handleCSV(file, header)
		if err != nil {
			http.Error(w, "error Parsing CSV\n", 500)
			log.Error().Err(err).Str(zerolog.CallerFieldName, "upload_web.HandleUpload()").Send()
			return
		}
	default: // archive? was uploaded
		payloader, err := handler.handleFile(file, header)
		if err != nil {
			http.Error(w, "error Handling File\n", 500)
			log.Error().Err(err).Str(zerolog.CallerFieldName, "upload_web.HandleUpload()").Send()
			return
		}
		payload = payloader.ToPayload()
	}

	if err := json.NewEncoder(w).Encode(payload); err != nil {
		http.Error(w, "error encoding JSON response\n", 500)
		log.Err(err).Send()
		return
	}
}

// handleFile passes the file to the UploadController for temporary storage
func (handler UploadHandler) handleFile(file multipart.File, header *multipart.FileHeader) (*HandledFile, error) {
	storedFile, hash, header, err := handler.UploadController.HandleFile(file, header)
	if err != nil {
		return nil, err
	}

	return &HandledFile{storedFile, hash, header}, err
}

// handleCSV passes a CSV file to the UploadController for loading of data
func (handler UploadHandler) handleCSV(file multipart.File, header *multipart.FileHeader) (*UploadPayload, error) {
	fileName, contentType, rawHeader, extra, err := handler.UploadController.HandleCSV(file, header)
	if err != nil {
		return nil, err
	}

	return &UploadPayload{
		Filename:    fileName,
		Uploadname:  fileName,
		Sha1:        "",
		ContentType: contentType,
		IsMeta:      true,
		RawHeader:   rawHeader,
		Extra:       extra,
	}, nil
}

// HandleProcessRequest is the handler for extracting and processing an archive.
// It receives an array of processRequestFile from the frontend.
// It then tries to load already known data on these archives.
//
// If any of the requested files encounter errors, a CSV of those errors is returned to the frontend.
// If no errors are encountered, a CSV containing the known data is returned to the frontend.
// Then in a background goroutine, the unknown archives are processed. See processUpload for more on this.
func (handler UploadHandler) HandleProcessRequest(w http.ResponseWriter, r *http.Request) {
	var body []processRequestFile
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "Error Parsing JSON Body", 500)
		log.Error().Err(err).Msg("error parsing JSON body")
		return
	}

	errorFiles := [][]string{{"File Name", "Stage", "ID", "Upload Name", "Content-Type"}}
	processedFiles := [][]string{{"Verification Code", "Sha256", "PkgFilename", "AssociatedLicense", "AssociatedLicenseRationale", "FamilyString"}}
	todo := make([]upload.Upload, 0, len(body))

	// fetch known data, and store errors and uploads needing processing
	for _, v := range body {
		u := upload.Upload{
			ID:          v.ID,
			Filename:    v.Filename,
			Uploadname:  v.Uploadname,
			ContentType: v.ContentType,
			Filepath:    v.Filepath,
		}

		currentRow := []string{
			"",                                       // file verification code
			"",                                       // Sha256
			u.Filename,                               // PkgFilename
			"",                                       // Associated license
			"",                                       // License Rationale
			fmt.Sprintf("file name: %s", u.Filename), // Family String
		}

		arch, fc, err := handler.UploadController.LookupArchive(u)
		if arch != nil && arch.Sha256.Valid {
			currentRow[1] = arch.Sha256.String
		}
		if err != nil {
			log.Error().Err(err).Interface("file", u).Msg("error looking up archive")

			errorFiles = append(errorFiles, []string{u.Filename, "LookupArchive", u.ID, u.Uploadname, u.ContentType})
			processedFiles = append(processedFiles, currentRow)
			continue
		}

		if fc != nil { // arch is also assumed to not be nil
			var verificationCode string
			if fc.VerificationCodeTwo != nil {
				verificationCode = hex.EncodeToString(fc.VerificationCodeTwo)
			} else if fc.VerificationCodeOne != nil {
				verificationCode = hex.EncodeToString(fc.VerificationCodeOne)
			}

			currentRow[0] = verificationCode
			currentRow[3] = fc.LicenseExpression.String
			currentRow[4] = fc.LicenseRationale.String
			currentRow[5] = fc.GroupName.String
			log.Trace().Str("GroupName", fc.GroupName.String).Msg("Setting Group Name")
			processedFiles = append(processedFiles, currentRow)

			continue
		} else { // archive needs processing
			todo = append(todo, u)
			processedFiles = append(processedFiles, currentRow)
			continue
		}
	}

	// send results of known data or errors to frontend
	if len(errorFiles) > 1 {
		log.Error().Int("body", len(body)).Int("errors", len(errorFiles)-1).Interface("error files", errorFiles).
			Msg("error processing files")

		w.Header().Set("Content-Type", "text/csv")
		w.WriteHeader(500)
		if err := csv.NewWriter(w).WriteAll(errorFiles); err != nil {
			http.Error(w, "error writing CSV Data", 500)
			log.Error().Err(err).Msg("CSV error")
		}
	} else {
		w.Header().Set("Content-Type", "text/csv")
		if err := csv.NewWriter(w).WriteAll(processedFiles); err != nil {
			http.Error(w, "error writing CSV Data", 500)
			log.Error().Err(err).Msg("CSV error")
		}
	}

	// process todo in the background
	if len(todo) > 0 {
		log.Info().Int("len(todo)", len(todo)).Msg("processing archives in the background")
		go func() {
			for _, v := range todo {
				if _, _, err := handler.UploadController.ProcessArchive(v, nil); err != nil {
					log.Error().Err(err).Interface("file", v).Msg("error processing archive")
					continue
				}
			}

			log.Info().Int("len(todo)", len(todo)).Msg("finished processing archives")
		}()
	}
}
