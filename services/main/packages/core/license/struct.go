// Copyright (c) 2020 Wind River Systems, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at:
//       http://www.apache.org/licenses/LICENSE-2.0
// Unless required by applicable law or agreed to in writing, software  distributed
// under the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES
// OR CONDITIONS OF ANY KIND, either express or implied.

package license

import (
	"database/sql"
	"wrs/tk/packages/core/archive"
	"wrs/tk/packages/core/file_collection"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

type License struct {
	LicenseID int64  `json:"license_id" db:"id"`
	Name      string `json:"license_name" db:"expression"`
	GroupID   int64  `json:"group_id,omitempty"`
	Group     string `json:"group,omitempty"`
}

type LicenseController struct {
	DB                       *sqlx.DB
	FileCollectionController file_collection.FileCollectionController
	ArchiveController        *archive.ArchiveController
}

func (controller LicenseController) GetByID(licenseID int64) (*License, error) {
	var ret License
	if err := controller.DB.QueryRowx("SELECT id, expression FROM license_expression WHERE id=$1",
		licenseID).Scan(&ret.LicenseID, &ret.Name); err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}

		return nil, err
	}

	return &ret, nil
}

func (controller LicenseController) GetByGroup(groupID int64) (*License, error) {
	ret := License{GroupID: groupID}
	var parentID sql.NullInt64

	if err := controller.DB.QueryRowx("SELECT name, associatedlicense, parent_id FROM group_container WHERE id=$1",
		groupID).Scan(&ret.Group, &ret.Name, &parentID); err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}

		return nil, err
	}

	if ret.Name != "" {
		return &ret, nil
	}

	if parentID.Valid { // Group has a parent, so check it for a license
		return controller.GetByGroup(parentID.Int64)
	}

	return nil, ErrNotFound
}

func (controller LicenseController) GetByFileCollection(verificationCode []byte, fileCollectionID int64) (*License, error) {
	container, err := controller.FileCollectionController.GetBy(verificationCode, fileCollectionID)
	if err != nil {
		if err == file_collection.ErrNotFound {
			return nil, ErrNotFound
		}

		return nil, err
	}

	if container.LicenseID.Valid {
		return controller.GetByID(container.LicenseID.Int64)
	}
	if container.GroupID.Valid {
		return controller.GetByGroup(container.GroupID.Int64)
	}

	return nil, ErrNotFound
}

func (controller LicenseController) GetByArchive(sha256 string, sha1 string, name string) (*License, error) {
	arch, err := controller.ArchiveController.GetBy(sha256, sha1, name)
	if err != nil {
		if err == archive.ErrNotFound {
			return nil, ErrNotFound
		}

		return nil, err
	}

	if arch.FileCollectionID.Valid {
		return controller.GetByFileCollection(nil, arch.FileCollectionID.Int64)
	}

	return nil, ErrNotFound
}

func (controller LicenseController) GetByContainer(verificationCode []byte, sha256 string, sha1 string, name string, fileCollectionID int64) (*License, error) {
	if license, err := controller.GetByFileCollection(verificationCode, fileCollectionID); err == nil {
		return license, nil
	} else if err != ErrNotFound {
		return nil, err
	}

	if license, err := controller.GetByArchive(sha256, sha1, name); err == nil {
		return license, nil
	} else if err != ErrNotFound {
		return nil, err
	}

	return nil, ErrNotFound
}

func (controller LicenseController) ParseLicenseExpression(expression string) (int64, error) {
	var lid int64
	if err := controller.DB.QueryRow("SELECT parse_license_expression($1)", expression).Scan(&lid); err != nil {
		return lid, errors.Wrapf(err, "error parsing license expression")
	}

	return lid, nil
}
