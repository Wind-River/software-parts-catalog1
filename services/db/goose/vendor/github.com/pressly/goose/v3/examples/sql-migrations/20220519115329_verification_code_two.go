package tkdb

import (
	"database/sql"
	"runtime"
	"wrs/tk/generics/graph"

	"gitlab.devstar.cloud/ip-systems/verification-code.git/code"

	"github.com/pkg/errors"
	"github.com/pressly/goose/v3"
	"github.com/rs/zerolog/log"
)

func init() {
	goose.AddMigration(upVerificationCodeTwo, downVerificationCodeTwo)

	_, filename, _, _ := runtime.Caller(0)
	log.Info().Str("up", "upVerificationCodeTwo").Str("down", "downVerificationCodeTwo").
		Str("filename", filename).Msg("Registered Migrations")
}

func upVerificationCodeTwo(tx *sql.Tx) error {
	// This code is executed when the migration is applied.
	if _, err := tx.Exec("ALTER TABLE file_collection " +
		"ADD COLUMN IF NOT EXISTS " +
		"verification_code_two BYTEA UNIQUE"); err != nil {
		return err
	}

	topRows, err := tx.Query("SELECT id FROM file_collection WHERE verification_code_two IS NULL")
	if err != nil {
		return err
	}
	defer topRows.Close()

	for topRows.Next() {
		var cid int64
		if err := topRows.Scan(&cid); err != nil {
			return err
		}

		// TODO calculate and set verification_code_two
		vcoder := code.NewVersionTwo().(*code.VersionTwoHasher)

		collectionGraph := graph.NewDirectedGraph[int64, int64]()
		root := collectionGraph.Insert(cid, cid, nil)

		// Generate graph
		if err := collectionGraph.TraverseUniqueEdges(func(id int64) error {
			currentNode := collectionGraph.Get(id)
			rows, err := tx.Query("SELECT child_id FROM file_collection_contains WHERE parent_id=$1", id)
			if err != nil {
				return errors.Wrapf(err, "error selecting file_collection's children")
			}
			defer rows.Close()

			for rows.Next() {
				var tmp int64
				if err := rows.Scan(&tmp); err != nil {
					return errors.Wrapf(err, "error scanning file_collection's children")
				}

				currentNode.Edges.Add(collectionGraph.Insert(tmp, tmp))
			}

			return nil
		}, root.ID); err != nil {
			return err
		}

		// Traverse graph and add all files
		if err := collectionGraph.TraverseUniqueEdges(func(id int64) error {
			// Select all files and feed to verification code
			rows, err := tx.Query("SELECT f.checksum_sha256 "+
				"FROM file_belongs_collection fbc "+
				"INNER JOIN file f ON f.id=fbc.file_id "+
				"WHERE fbc.file_collection_id=$1", id)
			if err != nil {
				return errors.Wrapf(err, "error selecting files of collection %d", id)
			}
			defer rows.Close()

			for rows.Next() {
				var tmpSha256 sql.NullString
				if err := rows.Scan(&tmpSha256); err != nil {
					return errors.Wrapf(err, "error scanning checksums of files of collection %d", id)
				}

				if !tmpSha256.Valid {
					return errors.New("scanned sha256 is nil")
				}

				if err := vcoder.AddSha256Hex(tmpSha256.String); err != nil {
					return err
				}
			}

			return nil
		}, root.ID); err != nil {
			return err
		}
	}

	return nil
}

func downVerificationCodeTwo(tx *sql.Tx) error {
	// This code is executed when the migration is rolled back.
	return nil
}
