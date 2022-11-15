// Copyright (c) 2020 Wind River Systems, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at:
//       http://www.apache.org/licenses/LICENSE-2.0
// Unless required by applicable law or agreed to in writing, software  distributed
// under the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES
// OR CONDITIONS OF ANY KIND, either express or implied.

package upload

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"time"
	"wrs/tk/packages/checksum/sha1"
	"wrs/tk/packages/checksum/sha256"

	"wrs/tk/packages/blob"
	"wrs/tk/packages/csvconv"

	"encoding/csv"

	"strings"

	"encoding/json"

	"encoding/hex"

	"wrs/tk/packages/core/archive"
	"wrs/tk/packages/core/file_collection"
	"wrs/tk/packages/core/license"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type Upload struct {
	ID          string `json:"id"`
	Filename    string `json:"filename"`
	Uploadname  string `json:"uploadname"`
	ContentType string `json:"contentType"`

	Filepath string
}

type UploadController struct {
	UploadDirectory    string
	tmpDirectory       string
	packageDirectory   string
	extractedDirectory string

	BlobStorage    blob.Storage
	CSVTransformer csvconv.CSVTransformer
	db             *sqlx.DB

	archiveController        *archive.ArchiveController
	fileCollectionController file_collection.FileCollectionController
	licenseController        license.LicenseController
}

func NewUploadController(uploadDirectory string, blobStorage blob.Storage, db *sqlx.DB,
	archiveController *archive.ArchiveController,
	fileCollectionController file_collection.FileCollectionController,
	licenseController license.LicenseController) (*UploadController, error) {
	var ret UploadController

	ret.BlobStorage = blobStorage
	ret.CSVTransformer = csvconv.CSVTransformer{ImplicitConversion: true}
	ret.db = db
	ret.archiveController = archiveController
	ret.fileCollectionController = fileCollectionController
	ret.licenseController = licenseController

	if uploadDirectory != "" {
		abs, err := filepath.Abs(uploadDirectory)
		if err == nil {
			ret.UploadDirectory = abs
		}
	}

	stat, err := os.Stat(ret.UploadDirectory)
	if err != nil {
		return nil, errors.Wrapf(err, "error statting upload directory")
	}

	if !stat.IsDir() {
		return nil, errors.Wrapf(err, "%s is not a directory", ret.UploadDirectory)
	}

	directories := []string{"tmp", "pkgs", "ext"}
	for _, v := range directories {
		if err := os.Mkdir(filepath.Join(ret.UploadDirectory, v), 0755); err != nil && !os.IsExist(err) {
			return nil, errors.Wrapf(err, "error making directory: %s", v)
		}
	}

	ret.tmpDirectory = filepath.Join(ret.UploadDirectory, "tmp")
	ret.packageDirectory = filepath.Join(ret.UploadDirectory, "pkgs")
	ret.extractedDirectory = filepath.Join(ret.UploadDirectory, "ext")
	return &ret, nil
}

func (controller *UploadController) HandleCSV(file multipart.File, header *multipart.FileHeader) (fileName string, contentType string, rawHeader *multipart.FileHeader, extra string, err error) {
	r := csv.NewReader(file)

	headRow, err := r.Read()
	if err != nil {
		return "", "", nil, "", errors.Wrap(err, "error reading head row")
	}

	if len(headRow) > 2 && headRow[0] == "Verification Code" && headRow[1] == "Sha256" {
		expectedColumns := len(headRow) // subsequent rows should be the same length
		errorRows := [][]string{{"Error", "Verification Code", "Sha256", "ExpectedColumns", "Row"}}

		vcodeIdx := 0
		shaIdx := 1
		// nameIdx := 2

		// find index of each field if it exists
		var licenseIdx, rationaleIdx, groupIdx int = -1, -1, -1
		for i, v := range headRow {
			switch v {
			// case "PkgFilename":
			// 	nameIdx = i
			case "AssociatedLicense":
				licenseIdx = i
			case "AssociatedLicenseRationale":
				rationaleIdx = i
			case "FamilyString":
				groupIdx = i
			}
		}

		// iterate csv rows
		i := -1
		for row, err := r.Read(); err != io.EOF; row, err = r.Read() {
			i++
			if err != nil {
				exc := errors.Wrap(err, "error reading csv row")
				errorRows = append(errorRows, append([]string{exc.Error()}, row...))
				continue
			}

			if len(row) != expectedColumns {
				exc := fmt.Errorf("number of columns does not match header")
				fmt.Fprintln(os.Stderr, exc.Error())

				if len(row) > 1 {
					rowJSON, _ := json.Marshal(row[3:])
					errorRows = append(errorRows, []string{exc.Error(), row[0], row[1], string(rowJSON)})
					continue
				}
			}

			vcode := row[vcodeIdx]
			sha := row[shaIdx]
			valueMap := map[string]interface{}{"vcode": nil, "sha": sha}

			var where string
			switch {
			case vcode != "" && len(vcode) == 40: // v0 file verification code
				rawSha1, err := hex.DecodeString(vcode)
				if err != nil {
					exc := fmt.Errorf("invalid v0 verification code")
					fmt.Fprintf(os.Stderr, exc.Error())

					rowJSON, _ := json.Marshal(row[2:])
					errorRows = append(errorRows, []string{exc.Error(), vcode, sha, string(rowJSON)})
					continue
				}
				valueMap["vcode"] = append([]byte("FVC1\000"), rawSha1...)
				where = "WHERE verification_code_one = :vcode"
			case vcode != "" && len(vcode) == 50 && vcode[0:10] == "4656433100": // v1 file verification code
				rawVerifacationCode, err := hex.DecodeString(vcode)
				if err != nil {
					exc := fmt.Errorf("invalid v1 verification code")
					fmt.Fprintf(os.Stderr, exc.Error())

					rowJSON, _ := json.Marshal(row[2:])
					errorRows = append(errorRows, []string{exc.Error(), vcode, sha, string(rowJSON)})
					continue

				}
				valueMap["vcode"] = rawVerifacationCode
				where = "WHERE verification_code_one = :vcode"
			case sha != "" && len(sha) == 64: // sha256
				where = "WHERE a.file_collection_id=c.id AND a.checksum_sha256 = :sha"
			default:
				exc := fmt.Errorf("both identifier fields were invalid")
				fmt.Fprintf(os.Stderr, exc.Error())

				rowJSON, _ := json.Marshal(row[2:])
				errorRows = append(errorRows, []string{exc.Error(), vcode, sha, string(rowJSON)})
				continue
			}

			set := make([]string, 0, 3)
			if licenseIdx != -1 {
				lid, err := controller.licenseController.ParseLicenseExpression(row[licenseIdx])
				if err != nil {
					exc := errors.Wrapf(err, "error parsing license expression")
					log.Error().Err(err).Msg("error parsing license expression")

					rowJSON, _ := json.Marshal(row[3:])
					errorRows = append(errorRows, append([]string{exc.Error(), vcode, sha, string(rowJSON)}))
					continue
				}
				fmt.Printf("license.ParseLicenseExpression(%s) -> %d\n", row[licenseIdx], lid)

				set = append(set, "license_id = :license")
				valueMap["license"] = lid
			}
			if rationaleIdx != -1 {
				set = append(set, "license_rationale = :rationale")
				valueMap["rationale"] = row[rationaleIdx]
			}
			if groupIdx != -1 {
				set = append(set, "group_container_id = (SELECT parse_group_path(:path))")
				valueMap["path"] = row[groupIdx]
			}

			var update string
			if fvc, ok := valueMap["vcode"]; ok && fvc != nil {
				update = fmt.Sprintf("UPDATE file_collection SET %s %s", strings.Join(set, ", "), where)
			} else {
				update = fmt.Sprintf("UPDATE file_collection c SET %s FROM archive a %s", strings.Join(set, ", "), where)
			}

			// fmt.Printf("update:\n\t%#+v\n\nvalueMap:\n\t%#+v\n", update, valueMap)
			log.Trace().Interface("row", row).
				Str("update", update).
				Interface("valueMap", valueMap).
				Msg("Loading data from CSV into database")
			res, err := controller.db.NamedExec(update, valueMap)
			if err != nil {
				exc := errors.Wrapf(err, "error updating file_collection")
				log.Error().Err(exc).Interface("valueMap", valueMap).Str("update query", update).Msg("error updating file_collection")

				rowJSON, _ := json.Marshal(row[2:])
				errorRows = append(errorRows, append([]string{exc.Error(), vcode, sha, string(rowJSON)}))
				continue
			} else if count, _ := res.RowsAffected(); count < 1 {
				exc := errors.New("Update affected 0 rows")
				log.Error().Err(exc).Int("row", i).Str("update", update).Interface("values", valueMap).Send()

				rowJSON, _ := json.Marshal(row[2:])
				errorRows = append(errorRows, append([]string{exc.Error(), vcode, sha, string(rowJSON)}))
				continue
			}
		}

		fileName = header.Filename
		contentType = header.Header.Get("Content-Type")
		rawHeader = header
		// payload := &UploadPayload{
		// 	header.Filename,
		// 	header.Filename,
		// 	"",
		// 	header.Header.Get("Content-Type"),
		// 	true,
		// 	header,
		// 	"",
		// }
		if len(errorRows) > 1 {
			var buf bytes.Buffer
			if err := csv.NewWriter(&buf).WriteAll(errorRows); err != nil {
				return fileName, contentType, rawHeader, "", errors.Wrapf(err, "error writing csv")
			}

			extra = buf.String()
		}
		return fileName, contentType, rawHeader, extra, nil
	} else {
		return "", "", header, "", errors.New(fmt.Sprintf("header Row not recognized: %v", headRow))
	}
}

func (controller *UploadController) HandleFile(file multipart.File, header *multipart.FileHeader) (*os.File, string, *multipart.FileHeader, error) {
	now := time.Now().UTC()
	name := fmt.Sprintf("%d+%s", now.Nanosecond(), header.Filename)

	outFile, err := os.Create(filepath.Join(controller.tmpDirectory, name))
	if err != nil {
		return nil, "", header, err
	}
	defer outFile.Close()

	out := bufio.NewWriter(outFile)
	defer out.Flush()

	data, err := io.ReadAll(file)
	if err != nil {
		return nil, "", header, err
	}
	buf := bytes.NewBuffer(data)

	hash, _ := sha1.BytesToSha1(data)

	_, err = buf.WriteTo(out)

	return outFile, hash, header, err
}

// LookupArchive attempts to lookup the archive and file collection of an upload.
// If the upload is unknown, it returns a Initialized Archive.
func (controller *UploadController) LookupArchive(u Upload) (*archive.Archive, *file_collection.FileCollection, error) {
	u.Filepath = filepath.Join(controller.tmpDirectory, u.Uploadname)

	shaTwo, err := sha256.Sha256OfFilePath(u.Filepath)
	if err != nil {
		return nil, nil, err
	}

	arch, err := controller.archiveController.GetBySha256(shaTwo)
	if arch == nil { // Manually collect additional fields
		arch, _ = archive.InitArchive(u.Filepath, u.Filename)
	} else {
		log.Debug().
			Str(zerolog.CallerFieldName, "*UploadController.LookupArchive()").
			Interface("arch", arch).
			Str("sha256", shaTwo).
			Msg("received from archiveController.GetBySha256")
	}
	if err == archive.ErrNotFound {
		return arch, nil, nil
	} else if err != nil {
		return arch, nil, err
	}

	if !arch.FileCollectionID.Valid {
		return arch, nil, nil
	}

	fileCollection, err := controller.fileCollectionController.GetByID(arch.FileCollectionID.Int64)
	if err == file_collection.ErrNotFound {
		return arch, nil, nil
	} else if err != nil {
		return arch, nil, err
	}

	return arch, fileCollection, nil
}

func (controller *UploadController) ProcessArchive(u Upload, arch *archive.Archive) (*archive.Archive, *file_collection.FileCollection, error) {
	u.Filepath = filepath.Join(controller.tmpDirectory, u.Uploadname)

	log.Debug().Str(zerolog.CallerFieldName, "*UploadController.ProcessArchive()").Interface("u", u).Interface("arch", arch).Send()

	var err error
	if arch == nil { // brand new archive
		arch, err = archive.InitArchive(u.Filepath, u.Filename)
		if err != nil {
			return arch, nil, err
		}
	} else {
		arch.Path.String = u.Filepath
		arch.Path.Valid = true
	}

	if !arch.FileCollectionID.Valid { // archive needs to be processed
		if err := controller.archiveController.Process(arch); err != nil {
			return arch, nil, err
		}
	}

	fileCollection, err := controller.fileCollectionController.GetByID(arch.FileCollectionID.Int64)
	if err != nil {
		return arch, nil, err
	}

	return arch, fileCollection, nil
}
