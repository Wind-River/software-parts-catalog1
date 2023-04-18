package model

import (
	"wrs/tk/packages/core/part"
)

// TODOING
type Part struct {
	ID                         part.ID  `json:"id"`
	Type                       string   `json:"type"`
	Name                       string   `json:"name"`
	Version                    string   `json:"version"`
	Label                      string   `json:"label"`
	FamilyName                 string   `json:"family_name"`
	FileVerificationCode       []byte   `json:"file_verification_code"`
	Size                       int64    `json:"size"`
	License                    *string  `json:"license"`
	LicenseRationale           *string  `json:"license_rationale"`
	Description                string   `json:"description"`
	Comprised                  *part.ID `json:"comprised"`
	LicenseNotice              *string  `json:"license_notice"`               // deprecated
	AutomationLicense          *string  `json:"automation_license"`           // deprecated
	AutomationLicenseRationale *string  `json:"automation_license_rationale"` // deprecated
}

func ToPart(p *part.Part) Part {
	ret := Part{
		ID:                   p.PartID,
		Type:                 p.Type.String,
		Name:                 p.Name.String,
		Version:              p.Version.String,
		Label:                p.Label.String,
		FamilyName:           p.FamilyName.String,
		FileVerificationCode: p.FileVerificationCode,
		Size:                 p.Size.Int64,
		License:              &p.License.String,
		LicenseRationale:     &p.LicenseRationale.String,
		Description:          p.Description.String,
		Comprised:            &p.Comprised,
	}

	if p.License.Valid {
		ret.License = &p.License.String
	}

	return ret
}
