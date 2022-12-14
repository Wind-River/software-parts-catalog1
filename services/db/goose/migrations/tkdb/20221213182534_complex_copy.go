package migrations

import (
	"database/sql"
	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigration(upComplexCopy, downComplexCopy)
}

func upComplexCopy(tx *sql.Tx) error {
	// This code is executed when the migration is applied.
	return nil
}

func downComplexCopy(tx *sql.Tx) error {
	// This code is executed when the migration is rolled back.
	return nil
}
