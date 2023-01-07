// Copyright (c) 2020 Wind River Systems, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at:
//       http://www.apache.org/licenses/LICENSE-2.0
// Unless required by applicable law or agreed to in writing, software  distributed
// under the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES
// OR CONDITIONS OF ANY KIND, either express or implied.

package file_collection

import (
	"database/sql"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"

	"gitlab.devstar.cloud/ip-systems/verification-code.git/code"
	"gitlab.devstar.cloud/ip-systems/verification-code.git/code/legacy"
)

type FileCollection struct {
	FileCollectionID    int64          `db:"id"`
	InsertDate          time.Time      `db:"insert_date"`
	GroupID             sql.NullInt64  `db:"group_container_id"`
	GroupName           sql.NullString `db:"group_name"`
	Extracted           bool           `db:"flag_extract"`
	LicenseExtracted    bool           `db:"flag_license_extracted"`
	LicenseID           sql.NullInt64  `db:"license_id"`
	LicenseRationale    sql.NullString `db:"license_rationale"`
	AnalystID           sql.NullInt64  `db:"analyst_id"`
	LicenseExpression   sql.NullString `db:"license_expression"`
	LicenseNotice       sql.NullString `db:"license_notice"`
	Copyright           sql.NullString `db:"copyright"`
	VerificationCodeOne []byte         `json:"verification_code_one,omitempty" db:"verification_code_one"`
	VerificationCodeTwo []byte         `json:"verification_code_two,omitempty" db:"verification_code_two"`
}

type FileCollectionController struct {
	DB *sqlx.DB
}

func (controller FileCollectionController) GetBy(verificationCode []byte, fileCollectionID int64) (*FileCollection, error) {
	var ret *FileCollection = nil
	var err error

	if verificationCode != nil && len(verificationCode) > 0 { // try fetching by verification code
		ret, err = controller.GetByVerificationCode(verificationCode)
		if err != nil && err != ErrNotFound {
			return nil, err
		}
	}

	if fileCollectionID > 0 { // try fetching by file collection id
		ret, err = controller.GetByID(fileCollectionID)
		if err != nil && err != ErrNotFound {
			return nil, err
		}
	}

	return ret, ErrNotFound
}

func (controller FileCollectionController) GetByVerificationCode(verificationCode []byte) (*FileCollection, error) {
	if len(verificationCode) == 0 {
		return nil, ErrNotFound
	}

	// upgrade verification code if necessary
	version, _ := code.VersionOf(verificationCode)
	if version != nil && *version == code.VERSION_ZERO {
		log.Debug().Str("verification_code", hex.EncodeToString(verificationCode)).Msg("upgrading version zero")
		var err error
		verificationCode, err = legacy.Upgrade(verificationCode)
		if err != nil {
			return nil, err
		}

		v := code.VERSION_ONE
		version = &v
	}
	var query string = "SELECT fc.*, build_group_path(fc.group_container_id) as group_name, " +
		"l.expression as license_expression " +
		"FROM file_collection AS fc " +
		"LEFT JOIN license_expression AS l ON l.id=fc.license_id "

	switch *version {
	case code.VERSION_ONE:
		query += "WHERE fc.verification_code_one=$1 "
	case code.VERSION_TWO:
		query += "WHERE fc.verification_code_two=$1 "
	default:
		return nil, errors.New(fmt.Sprintf("unsupported version: %v\n", version))
	}

	var ret FileCollection
	if err := controller.DB.QueryRowx(query, verificationCode).StructScan(&ret); err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}

		return nil, err
	}

	return &ret, nil
}

func (controller FileCollectionController) GetByID(fileCollectionID int64) (*FileCollection, error) {
	if fileCollectionID <= 0 {
		return nil, ErrNotFound
	}

	var ret FileCollection
	if err := controller.DB.QueryRowx("SELECT fc.*, build_group_path(fc.group_container_id) as group_name, "+
		"l.expression as license_expression "+
		"FROM file_collection AS fc "+
		"LEFT JOIN license_expression AS l ON l.id=fc.license_id "+
		"WHERE fc.id=$1",
		fileCollectionID).StructScan(&ret); err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}

		return nil, err
	}

	return &ret, nil
}

func (controller FileCollectionController) ShallowCountFiles(fileCollectionID int64) (int64, error) {
	var ret int64
	if err := controller.DB.QueryRowx("SELECT COUNT(fbc.*) FROM file_belongs_collection fbc WHERE fbc.file_collection_id=$1",
		fileCollectionID).Scan(&ret); err != nil {
		return ret, errors.Wrapf(err, "error counting files")
	}
	return ret, nil
}

func (controller FileCollectionController) CountFiles(fileCollectionID int64) (int64, error) {
	var ret int64
	if err := controller.DB.QueryRowx("SELECT COUNT(*) FROM select_file_collection_files($1)",
		fileCollectionID).Scan(&ret); err != nil {
		return ret, errors.Wrapf(err, "error counting files")
	}
	return ret, nil
}

func (controller FileCollectionController) CountSubCollections(fileCollectionID int64) (int64, error) {
	var ret int64
	if err := controller.DB.QueryRowx("SELECT COUNT(*) FROM file_collection_contains WHERE parent_id=$1",
		fileCollectionID).Scan(&ret); err != nil {
		return ret, errors.Wrapf(err, "error counting files")
	}
	return ret, nil
}

func (controller FileCollectionController) UpdateTribalKnowledge(fileCollectionID int64, licenseID int64, licenseRationale string, familyPath string) error {
	// Sanity check input
	if licenseID > 0 && licenseRationale == "" && familyPath == "" {
		return errors.New("no data was given to update")
	}

	valueMap := make(map[string]interface{})
	valueMap["fcid"] = fileCollectionID
	setFragments := make([]string, 0, 3)
	if licenseID > 0 {
		setFragments = append(setFragments, "license_id=:license")
		valueMap["license"] = licenseID
	}
	if licenseRationale != "" {
		setFragments = append(setFragments, "license_rationale=:rationale")
		valueMap["rationale"] = licenseRationale
	}
	if familyPath != "" {
		setFragments = append(setFragments, "group_container_id=(SELECT parse_group_path(:path))")
		valueMap["path"] = familyPath
	}

	sql := fmt.Sprintf("UPDATE file_collection SET %s WHERE id=:fcid RETURNING group_container_id", strings.Join(setFragments, ", "))
	log.Debug().Str("sql", sql).Interface("value_map", valueMap).Msg("Updating file_collection")

	res, err := controller.DB.NamedExec(sql, valueMap)
	if err != nil {
		log.Error().Err(err).Interface("value_map", valueMap).Str("sql", sql).Msg("error updating file_collection")
		return err
	} else if count, _ := res.RowsAffected(); count < 1 {
		log.Error().Err(err).Interface("value_map", valueMap).Str("sql", sql).Msg("update had no affect")
		return errors.New("update had no affect")
	}

	return nil
}
