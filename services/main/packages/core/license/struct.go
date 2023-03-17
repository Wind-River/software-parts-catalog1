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
	"wrs/tk/packages/core/archive"
	"wrs/tk/packages/core/part"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

type License struct {
	Name string `json:"license_name" db:"expression"`
}

type LicenseController struct {
	DB                *sqlx.DB
	PartController    part.PartController
	ArchiveController *archive.ArchiveController
}

// TODO should this be removed entirely?
func (controller LicenseController) GetByLicenseExpression(expression string) (*License, error) {
	var ret License
	ret.Name = expression

	// TODO check that a part has this license?

	return &ret, nil
}

func (controller LicenseController) GetByPart(verificationCode []byte, partID *part.ID) (*License, error) {
	container, err := controller.PartController.GetBy(verificationCode, partID)
	if err != nil {
		if err == part.ErrNotFound {
			return nil, ErrNotFound
		}

		return nil, err
	}
	if container.License.Valid {
		return controller.GetByLicenseExpression(container.License.String)
	}

	return nil, ErrNotFound
}

func (controller LicenseController) GetByArchive(sha256 []byte, sha1 []byte, name string) (*License, error) {
	arch, err := controller.ArchiveController.GetBy(sha256, sha1, name)
	if err != nil {
		if err == archive.ErrNotFound {
			return nil, ErrNotFound
		}

		return nil, err
	}

	if arch.PartID != nil {
		return controller.GetByPart(nil, arch.PartID)
	}

	return nil, ErrNotFound
}

func (controller LicenseController) GetByContainer(verificationCode []byte, sha256 []byte, sha1 []byte, name string, partID *part.ID) (*License, error) {
	if license, err := controller.GetByPart(verificationCode, partID); err == nil {
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
