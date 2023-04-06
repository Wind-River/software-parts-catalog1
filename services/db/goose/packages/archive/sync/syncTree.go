package sync

import (
	"database/sql"
	"path/filepath"
	"strings"
	"wrs/tkdb/goose/packages/archive/tree"
	"wrs/tkdb/goose/packages/part"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

func SyncTree(db *sql.Tx, partController *part.PartController, root tree.Node) (uuid.UUID, error) {
	prt, err := partController.GetByVerificationCode(root.GetFileVerificationCode())
	if err == nil {
		rootArchive, ok := root.(*tree.Archive)
		if ok {
			// upsert archive and archive_alias
			if _, err := db.Exec(`INSERT INTO archive (sha256, archive_size, md5, sha1, part_id) VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (sha256) DO UPDATE SET part_id=EXCLUDED.part_id`,
				rootArchive.Sha256[:], rootArchive.Size, rootArchive.Md5[:], rootArchive.Sha1[:], prt.PartID); err != nil {
				return uuid.Nil, errors.Wrapf(err, "error upserting archive")
			}

			if _, err := db.Exec(`INSERT INTO archive_alias (archive_sha256, name) VALUES ($1, $2) ON CONFLICT (archive_sha256, name) DO NOTHING`,
				rootArchive.Sha256[:], root.GetName()); err != nil {
				return uuid.Nil, errors.Wrapf(err, "error upserting archive_alias")
			}
		}

		return uuid.UUID(prt.PartID), nil
	} else if err == part.ErrNotFound {
		// Insert entire tree
		return syncTree(db, partController, root)
	} else { // unexpected error
		return uuid.Nil, err
	}
}

// trimPath is used to remove the first directory in the path
// This directory is specific to our extraction process, and is not a directory originally of the archive
func trimPath(path string) string {
	if path == "" {
		return ""
	}

	if index := strings.Index(path, "/"); index == -1 {
		return path
	}

	return filepath.Join(strings.Split(path, "/")[1:]...)
}

func syncTree(db *sql.Tx, partController *part.PartController, root tree.Node) (uuid.UUID, error) {
	// Insert part
	var partID uuid.UUID
	if err := db.QueryRow(`INSERT INTO part (type, name, license, license_rationale, license_notice)
	VALUES ('archive', $1, $2, $3, $4) RETURNING part_id`,
		root.GetName(), root.GetLicense(), root.GetLicenseRationale(), root.GetLicenseNotice()).Scan(&partID); err != nil {
		return partID, errors.Wrapf(err, "error creating part for archive")
	}

	// Upsert all files and file_aliases
	for _, subFile := range root.GetFiles() {
		if _, err := db.Exec(`INSERT INTO file (sha256, file_size, md5, sha1) VALUES ($1, $2, $3, $4) ON CONFLICT (sha256) DO NOTHING`,
			subFile.Sha256[:], subFile.Size, subFile.Md5[:], subFile.Sha1[:]); err != nil {
			return partID, errors.Wrapf(err, "error inserting file")
		}

		if _, err := db.Exec(`INSERT INTO file_alias (file_sha256, name) VALUES ($1, $2) ON CONFLICT (file_sha256, name) DO NOTHING`,
			subFile.Sha256[:], subFile.GetName()); err != nil {
			return partID, errors.Wrapf(err, "error inserting file_alias")
		}

		if _, err := db.Exec(`INSERT INTO part_has_file (part_id, file_sha256, path) VALUES ($1, $2, $3)`,
			partID, subFile.Sha256[:], trimPath(subFile.GetPath())); err != nil {
			log.Error().Err(err).Str("part_id", partID.String()).Hex("file_sha256", subFile.Sha256[:]).Hex("file_sha1", subFile.Sha1[:]).Str("path", subFile.GetPath()).
				Interface("file", *subFile.File).Msg("Added duplicate file?")
			return partID, errors.Wrapf(err, "error adding file to part (%s, [SHA256:%x SHA1:%x], %s)", partID.String(), subFile.Sha256, subFile.Sha1, subFile.GetPath())
		}
	}

	// Recursively sync all sub-archives
	for _, subNode := range root.GetNodes() {
		subPartID, err := SyncTree(db, partController, subNode.Node)
		if err != nil {
			return partID, errors.Wrapf(err, "error syncing sub-archive")
		}

		if _, err := db.Exec(`INSERT INTO part_has_part (parent_id, child_id, path) 
		VALUES ($1, $2, $3)`,
			partID, subPartID, subNode.Path); err != nil {
			return partID, errors.Wrapf(err, "error adding sub-archive (%s, %s, %s)", partID, subPartID, subNode.Path)
		}
	}

	// set file_verification_code
	if result, err := db.Exec(`UPDATE part SET file_verification_code=$1 WHERE part_id=$2`,
		root.GetFileVerificationCode(), partID); err != nil {
		return partID, errors.Wrapf(err, "error updating file_verification_code of part: \"%s\"", partID.String())
	} else {
		count, err := result.RowsAffected()
		if err != nil {
			return partID, errors.Wrapf(err, "error checking result of setting file_verification_code")
		}
		if count != 1 {
			return partID, errors.Wrapf(err, "setting file_verification_code of %s to %x affected %d rows", partID, root.GetFileVerificationCode(), count)
		}
	}

	if rootArchive, ok := root.(*tree.Archive); ok {
		// Insert archive and archive_alias
		if _, err := db.Exec(`INSERT INTO archive (sha256, archive_size, md5, sha1, part_id) VALUES ($1, $2, $3, $4, $5)
	ON CONFLICT (sha256) DO UPDATE SET part_id=EXCLUDED.part_id`,
			rootArchive.Sha256[:], rootArchive.Size, rootArchive.Md5[:], rootArchive.Sha1[:], partID); err != nil {
			return partID, errors.Wrapf(err, "error inserting root archive")
		}

		if _, err := db.Exec(`INSERT INTO archive_alias (archive_sha256, name) VALUES ($1, $2) ON CONFLICT (archive_sha256, name) DO NOTHING`,
			rootArchive.Sha256[:], root.GetName()); err != nil {
			return partID, errors.Wrapf(err, "error inserting root archive alias")
		}
	}

	// Insert duplicates
	if len(root.GetDuplicates()) > 0 {
		for _, v := range root.GetDuplicates() {
			if archive, ok := v.(*tree.Archive); ok {
				// Insert archive and archive_alias
				if _, err := db.Exec(`INSERT INTO archive (sha256, archive_size, md5, sha1, part_id) VALUES ($1, $2, $3, $4, $5)
	ON CONFLICT (sha256) DO UPDATE SET part_id=EXCLUDED.part_id`,
					archive.Sha256[:], archive.Size, archive.Md5[:], archive.Sha1[:], partID); err != nil {
					return partID, errors.Wrapf(err, "error inserting root archive")
				}

				if _, err := db.Exec(`INSERT INTO archive_alias (archive_sha256, name) VALUES ($1, $2) ON CONFLICT (archive_sha256, name) DO NOTHING`,
					archive.Sha256[:], v.GetName()); err != nil {
					return partID, errors.Wrapf(err, "error inserting root archive alias")
				}
			}
		}
	}

	return partID, nil
}
