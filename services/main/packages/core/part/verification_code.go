package part

import (
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"gitlab.devstar.cloud/ip-systems/verification-code.git/code"
)

// processFileCollection catalogs files found at parentDirectory as children of this archive, and recursively processes any sub-packages.
// if any file is missing a sha256, the resulting file verification code 2 will be nil
func (p *PartController) CalculateFileCollectionVerificationCode(partID uuid.UUID) (vcodeTwo []byte, err error) {
	vcoderTwo := code.NewVersionTwo().(*code.VersionTwoHasher)

	// Select all files and feed to verification code
	rows, err := p.DB.Queryx("SELECT f.checksum_sha256 "+
		"FROM part_has_file phf "+
		"INNER JOIN file f ON f.sha256=phf.phf.file_sha256 "+
		"WHERE phf.part_id=$1", partID)
	if err != nil {
		return nil, errors.Wrapf(err, "error selecting files of part %s", partID)
	}
	defer rows.Close()

	for rows.Next() {
		var tmpSha256 []byte
		if err := rows.Scan(&tmpSha256); err != nil {
			return nil, errors.Wrapf(err, "error scanning checksum of files of part %s", partID)
		}

		if err := vcoderTwo.AddSha256(tmpSha256); err != nil {
			return nil, err
		}
	}
	rows.Close()

	pgraph, err := NewPartGraph(p.DB, partID)
	if err != nil {
		return nil, err
	}

	if len(pgraph.Edges) > 0 {
		if err := pgraph.TraverseUniqueEdges(func(partID string) error {
			// Select all files and feed to verification code
			rows, err = p.DB.Queryx("SELECT f.sha256 "+
				"FROM part_has_file phf "+
				"INNER JOIN file f ON f.sha256=phf.file_sha256 "+
				"WHERE phf.part_id=$1", partID)
			if err != nil {
				return errors.Wrapf(err, "error selecting files of part %s", partID)
			}
			defer rows.Close()

			for rows.Next() {
				var tmpSha256 []byte
				if err := rows.Scan(&tmpSha256); err != nil {
					return errors.Wrapf(err, "error scanning checksums of files of part %s", partID)
				}

				if err := vcoderTwo.AddSha256(tmpSha256); err != nil {
					return err
				}
			}

			return nil
		}); err != nil {
			return nil, err
		}
	}

	vcodeTwo = vcoderTwo.Sum()
	return vcodeTwo, nil
}
