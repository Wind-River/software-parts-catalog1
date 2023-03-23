package model

import (
	"encoding/json"
	"wrs/tk/packages/core/part"
)

type Part struct {
	ID                         part.ID         `json:"id"`
	Type                       string          `json:"type"`
	Name                       string          `json:"name"`
	Version                    string          `json:"version"`
	FamilyName                 string          `json:"family_name"`
	FileVerificationCode       []byte          `json:"file_verification_code"`
	Size                       int64           `json:"size"`
	License                    *string         `json:"license"`
	LicenseRationale           json.RawMessage `json:"license_rationale"`
	LicenseNotice              string          `json:"license_notice"`
	AutomationLicense          string          `json:"automation_license"`
	AutomationLicenseRationale json.RawMessage `json:"automation_license_rationale"`
	Comprised                  *part.ID        `json:"comprised"`
}

func ToPart(p *part.Part) Part {
	ret := Part{
		ID:                   p.PartID,
		Type:                 p.Type.String,
		Name:                 p.Name.String,
		Version:              p.Version.String,
		FamilyName:           p.FamilyName.String,
		FileVerificationCode: p.FileVerificationCode,
		Size:                 p.Size.Int64,
		// License:                    p.License.String,
		LicenseRationale:           json.RawMessage(p.LicenseRationale.String),
		LicenseNotice:              p.LicenseNotice.String,
		AutomationLicense:          p.AutomationLicense.String,
		AutomationLicenseRationale: json.RawMessage(p.AutomationLicenseRationale.String),
		Comprised:                  &p.Comprised,
	}

	if p.License.Valid {
		ret.License = &p.License.String
	}

	return ret
}
