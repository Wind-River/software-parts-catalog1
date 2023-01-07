package file_collection

import (
	"database/sql"

	"github.com/pkg/errors"
	"gitlab.devstar.cloud/ip-systems/verification-code.git/code"
)

// processFileCollection catalogs files found at parentDirectory as children of this archive, and recursively processes any sub-packages.
// if any file is missing a sha256, the resulting file verification code 2 will be nil
func (p *FileCollectionController) CalculateFileCollectionVerificationCode(fileCollectionID int64) (vcodeOne []byte, vcodeTwo []byte, err error) {
	vcoderOne := code.NewVersionOne().(*code.VersionOneHasher)
	vcoderTwo := code.NewVersionTwo().(*code.VersionTwoHasher)

	// Select all files and feed to verification code
	rows, err := p.DB.Queryx("SELECT f.checksum_sha1, f.checksum_sha256 "+
		"FROM file_belongs_collection fbc "+
		"INNER JOIN file f ON f.id=fbc.file_id "+
		"WHERE fbc.file_collection_id=$1 AND flag_symlink=0 AND flag_fifo=0", fileCollectionID)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "error selecting files of file_collection %d", fileCollectionID)
	}
	defer rows.Close()

	for rows.Next() {
		var tmpSha1 sql.NullString
		var tmpSha256 sql.NullString
		if err := rows.Scan(&tmpSha1, &tmpSha256); err != nil {
			return nil, nil, errors.Wrapf(err, "error scanning checksum of files of file_collectior %d", fileCollectionID)
		}

		if err := vcoderOne.AddSha1Hex(tmpSha1.String); err != nil {
			return nil, nil, err
		}
		if vcoderTwo != nil {
			if !tmpSha256.Valid {
				vcoderTwo = nil
			} else {
				if err := vcoderTwo.AddSha256Hex(tmpSha256.String); err != nil {
					return nil, nil, err
				}
			}
		}
	}
	rows.Close()

	fcg, err := NewFileCollectionGraph(p.DB, fileCollectionID)
	if err != nil {
		return nil, nil, err
	}

	if len(fcg.Edges) > 0 {
		if err := fcg.TraverseUniqueEdges(func(collectionID int64) error {
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

				if err := vcoderOne.AddSha1Hex(tmpSha1.String); err != nil {
					return err
				}
				if vcoderTwo != nil {
					if !tmpSha256.Valid {
						vcoderTwo = nil
					} else {
						if err := vcoderTwo.AddSha256Hex(tmpSha256.String); err != nil {
							return err
						}
					}
				}
			}

			return nil
		}); err != nil {
			return nil, nil, err
		}
	}

	fvcOne := vcoderOne.Sum()
	var fvcTwo []byte
	if vcoderTwo != nil {
		fvcTwo = vcoderTwo.Sum()
	}
	return fvcOne, fvcTwo, nil
}
