//go:build tkdb

package tkdb

import (
	"database/sql"
	"time"

	"wrs/tkdb/goose/packages/archive/processor"
	"wrs/tkdb/goose/packages/archive/sync"
	"wrs/tkdb/goose/packages/archive/tree"
	"wrs/tkdb/goose/packages/part"

	"github.com/pkg/errors"
	"github.com/pressly/goose/v3"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func init() {
	goose.AddMigration(upComplexCopy, downComplexCopy)
}

func upComplexCopy(tx *sql.Tx) error {
	// This code is executed when the migration is applied.

	// oldFileCollections, err := collectSpecificCollectionsToTest(tx)
	oldFileCollections, err := collectCollectionsWithLicenseData(tx)
	if err != nil {
		return err
	}

	archiveProcessor, err := processor.NewArchiveProcessor(tx, nil, nil)
	if err != nil {
		return err
	}

	for _, v := range oldFileCollections {
		if err := processCollectionsArchives(archiveProcessor, v); err == processor.ErrSha256 {
			log.Warn().Int64("old file_collection_id", v.FileCollectionID).Msg("Skipping File Collection missing sha256 files")
		} else if err != nil {
			return err
		}

		archiveProcessor.Reset()
	}

	return nil
}

func downComplexCopy(tx *sql.Tx) error {
	// This code is executed when the migration is rolled back.
	return nil
}

// Used in place of collectCollectionsWithLicenseData to debug specific collections
func collectSpecificCollectionsToTest(tx *sql.Tx) ([]OldFilecollection, error) {
	oldCollections := make([]OldFilecollection, 0)

	rows, err := tx.Query(`SELECT fc.id, fc.insert_date, fc.flag_extract, fc.flag_license_extracted, fc.license_rationale, fc.verification_code_one, fc.verification_code_two,
		le.expression
		FROM file_collection fc
		INNER JOIN license_expression le ON le.id=fc.license_id
		WHERE fc.license_id IS NOT NULL
		AND fc.id IN (269056)`)
	// AND (SELECT COUNT(*) FROM archive_table WHERE file_collection_id=fc.id) > 0`)
	// AND fc.id NOT IN (6081, 6008)`) // TODO properly filter out collections with no archives
	if err != nil {
		return nil, errors.Wrapf(err, "error selecting file_collections")
	}
	defer rows.Close()

	for rows.Next() {
		var tmp OldFilecollection
		var licenseRationale sql.NullString
		var flagExtract, flagLicenseExtracted int
		if err := rows.Scan(&tmp.FileCollectionID, &tmp.InsertDate, &flagExtract,
			&flagLicenseExtracted, &licenseRationale,
			&tmp.FileVerificationCodeOne, &tmp.FileVerificationCodeTwo,
			&tmp.LicenseExpression); err != nil {
			return nil, errors.Wrapf(err, "error scanning file_collections")
		}

		if licenseRationale.Valid {
			tmp.LicenseRationale = licenseRationale.String
		}
		tmp.Extracted = flagExtract > 0
		tmp.LicenseExtracted = flagLicenseExtracted > 0

		oldCollections = append(oldCollections, tmp)
	}
	rows.Close()

	return oldCollections, nil
}

// Collect collections that have data to migrate
func collectCollectionsWithLicenseData(tx *sql.Tx) ([]OldFilecollection, error) {
	oldCollections := make([]OldFilecollection, 0)

	rows, err := tx.Query(`SELECT fc.id, fc.insert_date, fc.flag_extract, fc.flag_license_extracted, fc.license_rationale, fc.verification_code_one, fc.verification_code_two,
		le.expression
		FROM file_collection fc
		INNER JOIN license_expression le ON le.id=fc.license_id
		WHERE fc.license_id IS NOT NULL`)
	// AND (SELECT COUNT(*) FROM archive_table WHERE file_collection_id=fc.id) > 0`)
	// AND fc.id NOT IN (6081, 6008)`) // TODO properly filter out collections with no archives
	if err != nil {
		return nil, errors.Wrapf(err, "error selecting file_collections")
	}
	defer rows.Close()

	for rows.Next() {
		var tmp OldFilecollection
		var licenseRationale sql.NullString
		var flagExtract, flagLicenseExtracted int
		if err := rows.Scan(&tmp.FileCollectionID, &tmp.InsertDate, &flagExtract,
			&flagLicenseExtracted, &licenseRationale,
			&tmp.FileVerificationCodeOne, &tmp.FileVerificationCodeTwo,
			&tmp.LicenseExpression); err != nil {
			return nil, errors.Wrapf(err, "error scanning file_collections")
		}

		if licenseRationale.Valid {
			tmp.LicenseRationale = licenseRationale.String
		}
		tmp.Extracted = flagExtract > 0
		tmp.LicenseExtracted = flagLicenseExtracted > 0

		oldCollections = append(oldCollections, tmp)
	}
	rows.Close()

	return oldCollections, nil
}

// Try to find archives for the given file collection and migrate to a part
func processCollectionsArchives(archiveProcessor *processor.ArchiveProcessor, ofc OldFilecollection) error {
	logger := log.With().Str(zerolog.CallerFieldName, "processCollectionsArchives").Int64("file_collection_id", ofc.FileCollectionID).Logger()
	logger.Debug().Msg("start")
	defer logger.Debug().Msg("end")

	// find archives
	archiveIDs := make([]int64, 0)
	rows, err := archiveProcessor.Tx.Query(`SELECT id FROM archive_table 
	WHERE file_collection_id=$1 AND checksum_sha1 IS NOT NULL AND extract_status<>-404`,
		ofc.FileCollectionID)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var tmp int64
		if err := rows.Scan(&tmp); err != nil {
			return errors.Wrapf(err, "error scanning archive of collection %d", ofc.FileCollectionID)
		}

		archiveIDs = append(archiveIDs, tmp)
	}
	logger.Debug().Interface("archiveIDs", archiveIDs).Msg("collected archives")

	var root tree.Node
	if len(archiveIDs) == 0 {
		// log.Warn().Int64("file_collection_id", ofc.FileCollectionID).Msg("Skipping Collection With No Archives")
		logger.Debug().Msg("processing file_collcetion")
		root, err = archiveProcessor.ProcessCollection(ofc.FileCollectionID)
		if err != nil {
			return err
		}
	} else {
		logger.Debug().Msg("processing archive")
		// process one
		root, err = archiveProcessor.ProcessArchive(archiveIDs[0], nil)
		if err != nil {
			return errors.Wrapf(err, "error processing archive %d", archiveIDs[0])
		}
	}
	logger.Debug().Msg("processed root node")

	// calculate verification code
	if err := tree.CalculateVerificationCodes(root); err != nil {
		return err
	}
	logger.Debug().Hex("file_verification_code", root.GetFileVerificationCode()).Msg("calculated verification codes")

	rootUUID, err := sync.SyncTree(archiveProcessor.Tx, &part.PartController{DB: archiveProcessor.Tx}, root)
	if err != nil {
		return err
	}
	logger = logger.With().Str("uuid", rootUUID.String()).Logger()
	logger.Debug().Msg("synced tree")

	logger.Debug().Interface("archiveIDs", archiveIDs).Msg("upserting other archives")
	// upsert other archives
	for _, v := range archiveIDs {
		a, err := processor.InitArchive(archiveProcessor.Tx, v)
		if err != nil {
			return err
		}

		// upsert archive and archive_alias
		if _, err := archiveProcessor.Tx.Exec(`INSERT INTO archive (sha256, archive_size, md5, sha1, part_id) VALUES ($1, $2, $3, $4, $5)
				ON CONFLICT (sha256) DO UPDATE SET part_id=EXCLUDED.part_id`,
			a.Sha256[:], a.Size, a.Md5[:], a.Sha1[:], rootUUID); err != nil {
			return errors.Wrapf(err, "error upserting archive")
		}

		if _, err := archiveProcessor.Tx.Exec(`INSERT INTO archive_alias (archive_sha256, name) VALUES ($1, $2) ON CONFLICT (archive_sha256, name) DO NOTHING`,
			a.Sha256[:], a.GetName()); err != nil {
			return errors.Wrapf(err, "error upserting archive_alias")
		}
	}
	logger.Debug().Msg("upserted other archives")

	return nil
}

// Represents file_collection in old model
// file_collection -> part
type OldFilecollection struct {
	FileCollectionID        int64     `db:"id"`
	InsertDate              time.Time `db:"insert_date"`
	groupContainerID        int       `db:"group_container_id"` // TODO group name instead of id
	Extracted               bool      `db:"flag_extract"`
	LicenseExtracted        bool      `db:"flag_license_extracted"`
	LicenseID               int64     `db:"license_id"`
	LicenseRationale        string    `db:"license_rationale"`
	analystID               int64     `db:"analyst_id"`
	LicenseExpression       string    `db:"license_expression"`
	LicenseNotice           string    `db:"license_notice"`
	Copyright               string    `db:"copyright"`
	FileVerificationCodeOne []byte    `db:"verification_code_one"`
	FileVerificationCodeTwo []byte    `db:"verification_code_two"`
}
