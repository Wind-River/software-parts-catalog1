package file_collection

import (
	"database/sql"
	"time"
)

type FileCollection struct {
	FileCollectionID    int64          `db:"id"`
	InsertDate          time.Time      `db:"insert_date"`
	GroupID             sql.NullInt64  `db:"group_container_id"`
	GroupName           sql.NullString `db:"group_name"`
	Extracted           bool           `db:"flag_extract"`
	LicenseExtracted    bool           `db:"flag_license_extracted"`
	LicenseID           sql.NullInt64  `db:"license_id"`
	LicenseRationale    sql.NullString `db:"license_rationale"`
	AnalystID           sql.NullInt64  `db:"analyst_id"`
	LicenseExpression   sql.NullString `db:"license_expression"`
	LicenseNotice       sql.NullString `db:"license_notice"`
	Copyright           sql.NullString `db:"copyright"`
	VerificationCodeOne []byte         `json:"verification_code_one,omitempty" db:"verification_code_one"`
	VerificationCodeTwo []byte         `json:"verification_code_two,omitempty" db:"verification_code_two"`
}
