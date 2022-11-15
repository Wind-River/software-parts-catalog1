package model

type Container struct {
	GroupContainerID    *int64   `json:"group_container_id"`
	GroupName           *string  `json:"group_name"`
	ArchiveID           *int64   `json:"archive_id"`
	Name                *string  `json:"name"`
	Path                *string  `json:"path"`
	Size                *int64   `json:"size"`
	Sha1                *string  `json:"sha1"`
	Sha256              *string  `json:"sha256"`
	Md5                 *string  `json:"md5"`
	FileCollectionID    *int64   `json:"file_collection_id"`
	VerificationCodeOne *string  `json:"verification_code_one"`
	VerificationCodeTwo *string  `json:"verification_code_two"`
	License             *License `json:"license"`
	LicenseID           *int64   `json:"-"`
	LicenseRationale    *string  `json:"license_rationale"`
	Extracted           *bool    `json:"extracted"`
	LicenseExtracted    *bool    `json:"license_extracted"`
}
