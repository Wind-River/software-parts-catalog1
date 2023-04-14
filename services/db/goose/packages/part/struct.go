// Copyright (c) 2020 Wind River Systems, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at:
//       http://www.apache.org/licenses/LICENSE-2.0
// Unless required by applicable law or agreed to in writing, software  distributed
// under the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES
// OR CONDITIONS OF ANY KIND, either express or implied.

package part

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"

	"gitlab.devstar.cloud/ip-systems/verification-code.git/code"
)

type ID uuid.UUID

func (id ID) String() string {
	return uuid.UUID(id).String()
}

// Scan implements database/sql.scanner interface
// if column is null, scan value as uuid.Nil
// otherwise parse the UUID string
func (id *ID) Scan(value interface{}) error {
	if value == nil {
		*id = ID(uuid.Nil)
		return nil
	}

	if sv, err := driver.String.ConvertValue(value); err == nil {
		if v, ok := sv.(string); ok {
			uuid, err := uuid.Parse(v)
			if err != nil {
				return err
			}

			*id = ID(uuid)
			return nil
		}
	} else {
		return err
	}

	return errors.New("failed to scan ID")
}

type Part struct {
	PartID                     ID             `db:"part_id"`
	Type                       sql.NullString `db:"type"`
	Name                       sql.NullString `db:"name"`
	Version                    sql.NullString `db:"version"`
	FamilyName                 sql.NullString `db:"family_name"`
	FileVerificationCode       []byte         `db:"file_verification_code"`
	Size                       sql.NullInt64  `db:"size"`
	License                    sql.NullString `db:"license"`
	LicenseRationale           sql.NullString `db:"license_rationale"`
	LicenseNotice              sql.NullString `db:"license_notice"`
	AutomationLicense          sql.NullString `db:"automation_license"`
	AutomationLicenseRationale sql.NullString `db:"automation_license_rationale"`
	Comprised                  ID             `db:"comprised"`
}

type PartController struct {
	DB *sql.Tx
}

func (controller PartController) GetBy(verificationCode []byte, partID *ID) (*Part, error) {
	var ret *Part = nil
	var err error

	if len(verificationCode) > 0 { // try fetching by verification code
		ret, err = controller.GetByVerificationCode(verificationCode)
		if err != nil && err != ErrNotFound {
			return nil, err
		}
	}

	if partID != nil { // try fetching by id
		ret, err = controller.GetByID(*partID)
		if err != nil && err != ErrNotFound {
			return nil, err
		}
	}

	return ret, ErrNotFound
}

func (controller PartController) GetByVerificationCode(verificationCode []byte) (*Part, error) {
	if len(verificationCode) == 0 {
		return nil, ErrNotFound
	}

	// check verification code version
	if version, _ := code.VersionOf(verificationCode); version != nil && *version != code.VERSION_TWO {
		return nil, errors.New("expected FVC2")
	}

	var ret Part
	if err := controller.DB.QueryRow(`SELECT
	part_id, type, name, version, family_name, 
	file_verification_code, size, license, license_rationale, license_notice, 
	automation_license, automation_license_rationale, 
	comprised
	FROM part WHERE file_verification_code=$1`, verificationCode).Scan(
		&ret.PartID,
		&ret.Type,
		&ret.Name,
		&ret.Version,
		&ret.FamilyName,
		&ret.FileVerificationCode,
		&ret.Size,
		&ret.License,
		&ret.LicenseRationale,
		&ret.LicenseNotice,
		&ret.AutomationLicense,
		&ret.AutomationLicenseRationale,
		&ret.Comprised,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}

		return nil, err
	}

	return &ret, nil
}

func (controller PartController) GetByID(partID ID) (*Part, error) {
	var ret Part
	if err := controller.DB.QueryRow(`SELECT 
	part_id, type, name, version, family_name, 
	file_verification_code, size, license, license_rationale, license_notice, 
	automation_license, automation_license_rationale, 
	comprised
	FROM part WHERE part_id=$1`,
		partID).Scan(
		&ret.PartID,
		&ret.Type,
		&ret.Name,
		&ret.Version,
		&ret.FamilyName,
		&ret.FileVerificationCode,
		&ret.Size,
		&ret.License,
		&ret.LicenseRationale,
		&ret.LicenseNotice,
		&ret.AutomationLicense,
		&ret.AutomationLicenseRationale,
		&ret.Comprised,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}

		log.Error().Err(err).Str("partID", partID.String()).Msg("Error getting part by id")
		return nil, err
	}

	return &ret, nil
}

func (controller PartController) GetByComprised(comprisedID ID) ([]Part, error) {
	partIDs := make([]ID, 0)
	rows, err := controller.DB.Query("SELECT part_id FROM part WHERE comprised=$1", comprisedID)
	if err == sql.ErrNoRows {
		return nil, nil
	} else if err != nil {
		return nil, errors.Wrapf(err, "error selecting parts by comprised")
	}
	defer rows.Close()

	for rows.Next() {
		var tmp ID
		if err := rows.Scan(&tmp); err != nil {
			return nil, errors.Wrapf(err, "error scanning parts by comprised")
		}

		partIDs = append(partIDs, tmp)
	}
	rows.Close()

	ret := make([]Part, 0, len(partIDs))
	for _, partID := range partIDs {
		tmp, err := controller.GetByID(partID)
		if err != nil {
			return ret, err
		}

		ret = append(ret, *tmp)
	}

	return ret, nil
}

// ShallowCountFiles returns the number of files owned directly by the given part
// Files owned by sub-parts are not counted
func (controller PartController) ShallowCountFiles(partID ID) (int64, error) {
	var ret int64
	if err := controller.DB.QueryRow("SELECT COUNT(*) FROM part_has_file WHERE part_id=$1",
		partID).Scan(&ret); err != nil {
		return ret, errors.Wrapf(err, "error counting files")
	}
	return ret, nil
}

// CountFiles recursively counts files from the given part and any sub-parts
func (controller PartController) CountFiles(partID ID) (int64, error) {
	ret, err := controller.ShallowCountFiles(partID)
	if err != nil {
		return ret, nil
	}

	subPartIDs := make([]ID, 0)
	rows, err := controller.DB.Query(`SELECT child_id FROM part_has_part WHERE parent_id=$1`,
		partID)
	if err != nil && err != sql.ErrNoRows {
		return ret, errors.Wrapf(err, "error selecting sub-parts of %s", partID.String())
	}
	defer rows.Close()

	for rows.Next() {
		var tmp ID
		if err := rows.Scan(&tmp); err != nil {
			return ret, errors.Wrapf(err, "error scanning sub-part of %s", partID.String())
		}

		subPartIDs = append(subPartIDs, tmp)
	}
	rows.Close()

	for _, v := range subPartIDs {
		subCount, err := controller.CountFiles(v)
		if err != nil {
			return ret, errors.Wrapf(err, "error counting files of sub-part %s of %s", v.String(), partID.String())
		}

		ret += subCount
	}

	return ret, nil
}

// CountSubParts counts the number of sub-parts directly owned by the given part
func (controller PartController) CountSubParts(partID ID) (int64, error) {
	var ret int64
	if err := controller.DB.QueryRow("SELECT COUNT(*) FROM part_has_part WHERE parent_id=$1",
		partID).Scan(&ret); err != nil {
		return ret, errors.Wrapf(err, "error counting sub-parts")
	}
	return ret, nil
}

// appendFragment appends a value to a field iff the given value is not nil, is not zero
// This is used since graphql is giving us optional values as pointers
// isZero is checked to make sure we don't overwrite actual data
func appendFragment[T any](column string,
	optional *T,
	isZero func(T) bool,
	setFragments []string, values []interface{}) ([]string, []interface{}) {
	if optional == nil { // end early, no value
		return setFragments, values
	}
	if isZero != nil && isZero(*optional) {
		return setFragments, values // end early, no value
	}

	values = append(values, *optional)
	setFragments = append(setFragments, fmt.Sprintf("%s=$%d", column, len(values)))

	return setFragments, values
}

// UpdateTribalKnowledge takes optional values we are receiving from GraphQL and updates them in the database iff they are not nil and not their zero value
// TODO, should this function be updated to allow to nil or zero values, or should that be a different function we add?
func (controller PartController) UpdateTribalKnowledge(partID ID, partType *string, name *string, version *string, familyName *string, fileVerificationCode []byte, license *string, licenseRationale json.RawMessage, license_notice *string, automationLicense *string, automationLicenseRational json.RawMessage, comprised *ID) error {
	values := []interface{}{partID}
	setFragments := make([]string, 0)
	setFragments, values = appendFragment[string]("type", partType, nil, setFragments, values)
	setFragments, values = appendFragment[string]("name", name, nil, setFragments, values)
	setFragments, values = appendFragment[string]("version", version, nil, setFragments, values)
	setFragments, values = appendFragment[string]("family_name", familyName, nil, setFragments, values)
	setFragments, values = appendFragment[[]byte]("file_verification_code", &fileVerificationCode,
		func(b []byte) bool { return len(b) == 0 },
		setFragments, values)
	setFragments, values = appendFragment[string]("license", license, nil, setFragments, values)
	setFragments, values = appendFragment[json.RawMessage]("license_rationale", &licenseRationale,
		func(rm json.RawMessage) bool { return rm == nil }, setFragments, values)
	setFragments, values = appendFragment[string]("automation_license", automationLicense, nil, setFragments, values)
	setFragments, values = appendFragment[json.RawMessage]("automation_license_rationale", &automationLicenseRational,
		func(rm json.RawMessage) bool { return rm == nil }, setFragments, values)
	setFragments, values = appendFragment[ID]("comprised", comprised, func(i ID) bool {
		return i == ID(uuid.Nil)
	}, setFragments, values)

	sql := fmt.Sprintf("UPDATE part SET %s WHERE part_id=$1", strings.Join(setFragments, ", "))
	log.Debug().Str("sql", sql).Interface("values", values).Msg("Updating file_collection")

	res, err := controller.DB.Exec(sql, values...)
	if err != nil {
		log.Error().Err(err).Interface("values", values).Str("sql", sql).Msg("error updating file_collection")
		return err
	} else if count, _ := res.RowsAffected(); count < 1 {
		log.Error().Err(err).Interface("values", values).Str("sql", sql).Msg("update had no affect")
		return errors.New("update had no affect")
	}

	return nil
}

// CreateAlias upserts an alias to a given part
// If the alias already exists for the given part
// If the alias exists and is associated with a different part, an error is returned
func (controller PartController) CreateAlias(partId ID, alias string) (*ID, error) {
	var aliasesPartID ID
	if err := controller.DB.QueryRow(`INSERT INTO part_alias (alias, part_id) VALUES ($1, $2)
	ON CONFLICT (alias) DO UPDATE SET alias=EXCLUDED.alias
	RETURNING part_id`, // meaningless update required for return
		alias, partId).Scan(&aliasesPartID); err != nil {
		return nil, errors.Wrapf(err, "error inserting part_alias")
	}

	if aliasesPartID.String() != partId.String() {
		return nil, errors.New("alias part id mismtach")
	}

	return &aliasesPartID, nil
}

// AttachDocument upserts a document into part_has_document or part_documents, depending on if a title is given
func (controller PartController) AttachDocument(partId ID, key string, title *string, document json.RawMessage) error {
	if title == nil || *title == "" {
		// no title, so insert into part_has_document
		if _, err := controller.DB.Exec(`INSERT INTO part_has_document(part_id, key, document) VALUES ($1, $2, $3)
		ON CONFLICT (part_id, key) DO UPDATE SET document=EXCLUDED.document`,
			partId, key, document); err != nil {
			return errors.Wrapf(err, "error inserting into part_has_document")
		}

		return nil
	}

	if _, err := controller.DB.Exec(`INSERT INTO part_documents(part_id, key, title, document) VALUES ($1, $2, $3, $4)
	ON CONFLICT (part_id, key, title) DO UPDATE SET document=EXCLUDED.document`); err != nil {
		return errors.Wrapf(err, "error inserting into part_documents")
	}

	return nil
}

type SubPart struct {
	ID   ID     `db:"child_id"`
	Path string `db:"path"`
}

// SubParts returns an array of path/part relationships of sub-parts of the given part
func (controller PartController) SubParts(partID ID) ([]SubPart, error) {
	rows, err := controller.DB.Query("SELECT child_id, path FROM part_has_part WHERE parent_id=$1", partID)
	if err == sql.ErrNoRows {
		return nil, nil
	} else if err != nil {
		return nil, errors.Wrapf(err, "error selecting sub-parts")
	}
	defer rows.Close()

	ret := make([]SubPart, 0)
	for rows.Next() {
		var id ID
		var path string
		if err := rows.Scan(&id, &path); err != nil {
			return nil, errors.Wrapf(err, "error scanning sub-parts")
		}

		ret = append(ret, SubPart{
			ID:   id,
			Path: path,
		})
	}

	return ret, nil
}

// AddPartToPart adds a sub-part to a part at a path
// If the relationship already exists, nothing changes
func (controller PartController) AddPartToPart(childID ID, parentID ID, path string) error {
	if _, err := controller.DB.Exec("INSERT INTO part_has_part (parent_id, child_id, path) VALUES ($1, $2, $3) ON CONFLICT (parent_id, child_id, path) DO NOTHING", // TODO this probably shouldn't catch the conflict, to make sure users didn't accidentally set the same path twice
		parentID, childID, path); err != nil {
		return errors.Wrapf(err, "error inserting part_has_part")
	}

	return nil
}
