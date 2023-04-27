package model

import (
	"fmt"
	"reflect"
	"strings"
	"unicode"
	"wrs/tk/packages/core/part"

	"github.com/pkg/errors"
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

// TypeToLTree converts a part type, which is styled like a file path, into a PostgreSQL ltree
// It also validates to make sure it is an accepted type.
func TypeToLTree(partType string) (string, error) {
	if partType == "" {
		return "", errors.New("empty part type")
	}

	// convert whole type to type components
	var word string
	components := make([]string, 0)
	for i, r := range partType {
		if r == '/' {
			if i == 0 {
				// ignore root '/'
				continue
			} else if word == "" {
				// ignore extraneous separator
				continue
			} else {
				components = append(components, word)
				word = ""
			}
		} else if unicode.IsLetter(r) || unicode.IsDigit(r) {
			word += string(r)
		} else {
			return "", errors.New("invalid part type")
		}
	}
	if word != "" {
		components = append(components, word)
	}

	if len(components) == 0 {
		return "", errors.New("empty part type")
	}

	// validate components
	root := components[0]
	subTypes := components[1:]
	switch root {
	case "file":
		switch {
		case len(subTypes) == 0: // file
			return "", errors.New("file expects a sub-type")
		case reflect.DeepEqual(subTypes, []string{"source"}): // file/source
		case reflect.DeepEqual(subTypes, []string{"binary"}): // file/binary
		case subTypes[0] == "custom" && len(subTypes) > 1: // file/custom/*
		default:
			return "", errors.New(fmt.Sprintf("unsupported sub-type for file: %s", strings.Join(subTypes, "/")))
		}
	case "archive":
		switch {
		case len(subTypes) == 0: // archive
		case subTypes[0] == "custom" && len(subTypes) > 1: // archive/custom/*
			fmt.Printf("subTypes: %#v\n", subTypes)
		default:
			return "", errors.New(fmt.Sprintf("unsupported sub-type for archive: %s", strings.Join(subTypes, "/")))
		}
	case "container":
		switch {
		case len(subTypes) == 0: // container
			return "", errors.New("container expects a sub-type")
		case reflect.DeepEqual(subTypes, []string{"image"}): // container/image
		case reflect.DeepEqual(subTypes, []string{"source"}): // container/source
		case subTypes[0] == "custom" && len(subTypes) > 1: // container/custom/*
		default:
			return "", errors.New(fmt.Sprintf("unsupported sub-type for container: %s", strings.Join(subTypes, "/")))
		}
	case "logical":
		switch {
		case len(subTypes) == 0: // logical
		case subTypes[0] == "custom" && len(subTypes) > 1: // logical/custom/*
		default:
			return "", errors.New(fmt.Sprintf("unsupported sub-type for logical: %s", strings.Join(subTypes, "/")))
		}
	default:
		return "", errors.New(fmt.Sprintf("unxpected root %s", components[0]))
	}

	return strings.Join(components, "."), nil
}

// Re-stylizes an ltree like a filepath
func LTreeToPath(lTree string) string {
	return "/" + strings.ReplaceAll(lTree, ".", "/")
}
