// Code generated by github.com/99designs/gqlgen, DO NOT EDIT.

package model

type ArchiveDistance struct {
	Distance int64    `json:"distance"`
	Archive  *Archive `json:"archive"`
}

type Document struct {
	Title    *string `json:"title"`
	Document Json    `json:"document"`
}

type NewPartInput struct {
	Type             *string `json:"type"`
	Name             *string `json:"name"`
	Version          *string `json:"version"`
	Label            *string `json:"label"`
	FamilyName       *string `json:"family_name"`
	License          *string `json:"license"`
	LicenseRationale *string `json:"license_rationale"`
	Description      *string `json:"description"`
	Comprised        *string `json:"comprised"`
}

type PartInput struct {
	ID                   string  `json:"id"`
	Type                 *string `json:"type"`
	Name                 *string `json:"name"`
	Version              *string `json:"version"`
	Label                *string `json:"label"`
	FamilyName           *string `json:"family_name"`
	FileVerificationCode *string `json:"file_verification_code"`
	License              *string `json:"license"`
	LicenseRationale     *string `json:"license_rationale"`
	Description          *string `json:"description"`
	Comprised            *string `json:"comprised"`
}

type Profile struct {
	Key       string      `json:"key"`
	Documents []*Document `json:"documents"`
}

type SearchCosts struct {
	Insert      int  `json:"insert"`
	Delete      int  `json:"delete"`
	Substitute  int  `json:"substitute"`
	MaxDistance *int `json:"max_distance"`
}

type SubPart struct {
	Path string `json:"path"`
	Part *Part  `json:"part"`
}

type UploadedArchive struct {
	Extracted bool     `json:"extracted"`
	Archive   *Archive `json:"archive"`
}
