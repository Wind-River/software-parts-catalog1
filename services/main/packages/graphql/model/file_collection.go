package model

import (
	"time"
	"wrs/tk/packages/core/file_collection"
)

type FileCollection struct {
	ID                int64     `json:"id"`
	InsertDate        time.Time `json:"insert_date"`
	GroupContainerID  *int64    `json:"group_container_id"`
	Extracted         bool      `json:"flag_extract"`
	LicenseExtracted  bool      `json:"flag_license_extract"`
	LicenseID         *int64    `json:"license_id"`
	LicenseRationale  string    `json:"license_rationale"`
	AnalystID         *int64    `json:"analyst_id"`
	LicenseExpression string    `json:"license_expression"`
	LicenseNotice     string    `json:"license_notice"`
	Copyright         string    `json:"copyright"`
	FVCOne            []byte    `json:"verification_code_one"`
	FVCTwo            []byte    `json:"verification_code_two"`
}

func ToFileCollection(fc *file_collection.FileCollection) FileCollection {
	ret := FileCollection{
		ID:         fc.FileCollectionID,
		InsertDate: fc.InsertDate,
		// GroupContainerID:  &fc.GroupID.Int64,
		Extracted:        fc.Extracted,
		LicenseExtracted: fc.LicenseExtracted,
		// LicenseID:         &fc.LicenseID.Int64,
		LicenseRationale: fc.LicenseRationale.String,
		// AnalystID:         &fc.AnalystID.Int64,
		LicenseExpression: fc.LicenseExpression.String,
		LicenseNotice:     fc.LicenseNotice.String,
		Copyright:         fc.Copyright.String,
		FVCOne:            fc.VerificationCodeOne,
		FVCTwo:            fc.VerificationCodeTwo,
	}

	if fc.GroupID.Valid {
		ret.GroupContainerID = &fc.GroupID.Int64
	}
	if fc.LicenseID.Valid {
		ret.LicenseID = &fc.LicenseID.Int64
	}
	if fc.AnalystID.Valid {
		ret.AnalystID = &fc.AnalystID.Int64
	}

	return ret
}
