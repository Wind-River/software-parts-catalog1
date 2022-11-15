// Copyright (c) 2020 Wind River Systems, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at:
//       http://www.apache.org/licenses/LICENSE-2.0
// Unless required by applicable law or agreed to in writing, software  distributed
// under the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES
// OR CONDITIONS OF ANY KIND, either express or implied.

package archive

import (
	"database/sql"

	"github.com/pkg/errors"
	"gitlab.devstar.cloud/ip-systems/verification-code.git/code"
)

// processFileCollection catalogs files found at parentDirectory as children of this archive, and recursively processes any sub-packages.
func (p *ArchiveController) CalculateArchiveVerificationCode(archiveID int64) (vcodeOne []byte, vcodeTwo []byte, err error) {
	vcoderOne := code.NewVersionOne().(*code.VersionOneHasher)
	vcoderTwo := code.NewVersionTwo().(*code.VersionTwoHasher)

	// Select all files and feed to verification code
	rows, err := p.DB.Queryx("SELECT f.checksum_sha1, f.checksum_sha256 "+
		"FROM file_belongs_archive fba "+
		"INNER JOIN file_alias fa ON fa.id=fba.file_id "+
		"INNER JOIN file f ON f.id=fa.file_id "+
		"WHERE fba.archive_id=$1 AND flag_symlink=0 AND flag_fifo=0", archiveID)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "error selecting files of archive %d", archiveID)
	}
	defer rows.Close()

	for rows.Next() {
		var tmpSha1 sql.NullString
		var tmpSha256 sql.NullString
		if err := rows.Scan(&tmpSha1, &tmpSha256); err != nil {
			return nil, nil, errors.Wrapf(err, "error scanning checksum of files of archive %d", archiveID)
		}

		if tmpSha1.Valid {
			if err := vcoderOne.AddSha1Hex(tmpSha1.String); err != nil {
				return nil, nil, err
			}
		}
		if tmpSha256.Valid {
			if err := vcoderTwo.AddSha256Hex(tmpSha256.String); err != nil {
				return nil, nil, err
			}
		}
	}
	rows.Close()

	ag, err := NewArchiveGraph(p.DB, archiveID)
	if err != nil {
		return nil, nil, err
	}

	if len(ag.Edges) > 0 {
		if err := ag.TraverseUniqueEdges(func(collectionID int64) error {
			// Select all files and feed to verification code
			rows, err = p.DB.Queryx("SELECT f.checksum_sha1, f.checksum_sha256 "+
				"FROM file_belongs_collection fbc "+
				"INNER JOIN file f ON f.id=fbc.file_id "+
				"WHERE fbc.file_collection_id=$1", collectionID)
			if err != nil {
				return errors.Wrapf(err, "error selecting files of collection %d", collectionID)
			}
			defer rows.Close()

			for rows.Next() {
				var tmpSha1 sql.NullString
				var tmpSha256 sql.NullString
				if err := rows.Scan(&tmpSha1, &tmpSha256); err != nil {
					return errors.Wrapf(err, "error scanning checksums of files of collection %d", collectionID)
				}

				if tmpSha1.Valid {
					if err := vcoderOne.AddSha1Hex(tmpSha1.String); err != nil {
						return err
					}
				}
				if tmpSha256.Valid {
					if err := vcoderTwo.AddSha256Hex(tmpSha256.String); err != nil {
						return err
					}
				}
			}

			return nil
		}); err != nil {
			return nil, nil, err
		}
	}

	return vcoderOne.Sum(), vcoderTwo.Sum(), nil
}
