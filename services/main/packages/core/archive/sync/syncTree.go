package sync

import (
	"wrs/tk/packages/core/archive/tree"
	"wrs/tk/packages/core/part"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

func SyncTree(db *sqlx.DB, partController *part.PartController, root *tree.Archive) (uuid.UUID, error) {
	prt, err := partController.GetByVerificationCode(root.FileVerificationCode)
	if err == nil {
		// upsert archive and archive_alias
		if _, err := db.Exec(`INSERT INTO archive (sha256, archive_size, md5, sha1, part_id) VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (sha256) DO UPDATE SET part_id=EXCLUDED.part_id`,
			root.Sha256[:], root.Size, root.Md5[:], root.Sha1[:], prt.PartID); err != nil {
			return uuid.Nil, errors.Wrapf(err, "error upserting archive")
		}

		if _, err := db.Exec(`INSERT INTO archive_alias (archive_sha256, name) VALUES ($1, $2) ON CONFLICT (archive_sha256, name) DO NOTHING`,
			root.Sha256[:], root.GetName()); err != nil {
			return uuid.Nil, errors.Wrapf(err, "error upserting archive_alias")
		}

		return uuid.UUID(prt.PartID), nil
	} else if err == part.ErrNotFound {
		// Insert entire tree
		return syncTree(db, partController, root)
	} else { // unexpected error
		return uuid.Nil, err
	}
}

func syncTree(db *sqlx.DB, partController *part.PartController, root *tree.Archive) (uuid.UUID, error) {
	// Insert part
	var partID uuid.UUID
	if err := db.QueryRowx(`INSERT INTO part (type, name) VALUES ('archive', $1) RETURNING part_id`,
		root.GetName()).Scan(&partID); err != nil {
		return partID, errors.Wrapf(err, "error creating part for archive")
	}

	// Upsert all files and file_aliases
	for _, subFile := range root.Files {
		if _, err := db.Exec(`INSERT INTO file (sha256, file_size, md5, sha1) VALUES ($1, $2, $3, $4) ON CONFLICT (sha256) DO NOTHING`,
			subFile.Sha256[:], subFile.Size, subFile.Md5[:], subFile.Sha1[:]); err != nil {
			return partID, errors.Wrapf(err, "error inserting file")
		}

		if _, err := db.Exec(`INSERT INTO file_alias (file_sha256, name) VALUES ($1, $2) ON CONFLICT (file_sha256, name) DO NOTHING`,
			subFile.Sha256[:], subFile.GetName()); err != nil {
			return partID, errors.Wrapf(err, "error inserting file_alias")
		}

		tmpUUID := uuid.New()
		if _, err := db.Exec(`INSERT INTO part_has_file (part_id, file_sha256, path) VALUES ($1, $2, $3)`,
			partID, subFile.Sha256[:], tmpUUID.String()+"/"+subFile.GetPath()); err != nil { // TODO trim // fuzzing path, TODO fix
			return partID, errors.Wrapf(err, "error adding file to part (%s, %x, %s)", partID.String(), subFile.Sha256, subFile.GetPath())
		}
	}

	// Recursively sync all sub-archives
	for _, subArchive := range root.Archives {
		subPartID, err := SyncTree(db, partController, subArchive.Archive)
		if err != nil {
			return partID, errors.Wrapf(err, "error syncing sub-archive")
		}

		if _, err := db.Exec(`INSERT INTO part_has_part (parent_id, child_id, path) 
		VALUES ($1, $2, $3)`,
			partID, subPartID, subArchive.Path); err != nil {
			return partID, errors.Wrapf(err, "error adding sub-archive")
		}
	}

	// set file_verification_code
	if result, err := db.Exec(`UPDATE part SET file_verification_code=$1 WHERE part_id=$2`,
		root.FileVerificationCode, partID); err != nil {
		return partID, errors.Wrapf(err, "error updating file_verification_code of part")
	} else {
		count, err := result.RowsAffected()
		if err != nil {
			return partID, errors.Wrapf(err, "error checking result of setting file_verification_code")
		}
		if count != 1 {
			return partID, errors.Wrapf(err, "setting file_verification_code of %s to %x affected %d rows", partID, root.FileVerificationCode, count)
		}
	}

	// Insert archive and archive_alias
	if _, err := db.Exec(`INSERT INTO archive (sha256, archive_size, md5, sha1, part_id) VALUES ($1, $2, $3, $4, $5)
	ON CONFLICT (sha256) DO UPDATE SET part_id=EXCLUDED.part_id`,
		root.Sha256[:], root.Size, root.Md5[:], root.Sha1[:], partID); err != nil {
		return partID, errors.Wrapf(err, "error inserting root archive")
	}

	if _, err := db.Exec(`INSERT INTO archive_alias (archive_sha256, name) VALUES ($1, $2) ON CONFLICT (archive_sha256, name) DO NOTHING`,
		root.Sha256[:], root.GetName()); err != nil {
		return partID, errors.Wrapf(err, "error inserting root archive alias")
	}

	// Insert duplicates
	if len(root.DuplicateArchives) > 0 {
		for _, v := range root.DuplicateArchives {
			// Insert archive and archive_alias
			if _, err := db.Exec(`INSERT INTO archive (sha256, archive_size, md5, sha1, part_id) VALUES ($1, $2, $3, $4, $5)
	ON CONFLICT (sha256) DO UPDATE SET part_id=EXCLUDED.part_id`,
				v.Sha256[:], v.Size, v.Md5[:], v.Sha1[:], partID); err != nil {
				return partID, errors.Wrapf(err, "error inserting root archive")
			}

			if _, err := db.Exec(`INSERT INTO archive_alias (archive_sha256, name) VALUES ($1, $2) ON CONFLICT (archive_sha256, name) DO NOTHING`,
				v.Sha256[:], v.GetName()); err != nil {
				return partID, errors.Wrapf(err, "error inserting root archive alias")
			}
		}
	}

	return partID, nil
}
