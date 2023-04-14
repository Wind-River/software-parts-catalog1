package part

import (
	"database/sql"
	"encoding/json"

	"github.com/pkg/errors"
)

type Profile struct {
	Key       string
	Documents []Document
}

type Document struct {
	Title    *string
	Document json.RawMessage
}

// GetProfile gets all documents for a given part and profile
func (controller PartController) GetProfile(partID ID, profile string) (*Profile, error) {
	var ret Profile
	ret.Key = profile

	var tmpDocument json.RawMessage
	if err := controller.DB.QueryRow("SELECT document FROM part_has_document WHERE part_id=$1 AND key=$2", partID, profile).Scan(&tmpDocument); err == sql.ErrNoRows {
		ret.Documents = make([]Document, 0)
	} else if err != nil {
		return nil, errors.Wrapf(err, "error selecting document")
	} else {
		ret.Documents = []Document{Document{Document: tmpDocument}}
	}

	rows, err := controller.DB.Query(`SELECT title, document FROM part_documents WHERE part_id=$1 AND key=$2`, partID, profile)
	if err == sql.ErrNoRows {
		return &ret, nil
	} else if err != nil {
		return nil, errors.Wrapf(err, "error selecting documents")
	}
	defer rows.Close()

	for rows.Next() {
		var tmpTitle string
		var tmpDocument json.RawMessage

		if err := rows.Scan(&tmpTitle, &tmpDocument); err != nil {
			return nil, errors.Wrapf(err, "error scanning documents")
		}

		ret.Documents = append(ret.Documents, Document{
			Title:    &tmpTitle,
			Document: tmpDocument,
		})
	}

	return &ret, nil
}

// GetProfiles gets all profiles, which is a key and list of documents, for a given part
func (controller PartController) GetProfiles(partID ID) ([]Profile, error) {
	profileMap := make(map[string]*Profile)

	// select part_has_document first
	rows, err := controller.DB.Query("SELECT key, document FROM part_has_document WHERE part_id=$1", partID)
	if err != nil && err != sql.ErrNoRows {
		return nil, errors.Wrapf(err, "error selecting every document")
	} else if err == nil {
		defer rows.Close()
		for rows.Next() {
			var tmpKey string
			var tmpDocument json.RawMessage

			if err := rows.Scan(&tmpKey, &tmpDocument); err != nil {
				return nil, errors.Wrapf(err, "error scanning every document")
			}

			profile, ok := profileMap[tmpKey]
			if !ok {
				profile = &Profile{
					Key:       tmpKey,
					Documents: []Document{{Document: tmpDocument}},
				}
			} else {
				profile.Documents = append(profile.Documents, Document{Document: tmpDocument})
			}

			profileMap[profile.Key] = profile
		}
		rows.Close()
	}
	// select part_documents second
	rows, err = controller.DB.Query("SELECT key, title, document FROM part_documents WHERE part_id=$1", partID)
	if err != nil && err != sql.ErrNoRows {
		return nil, errors.Wrapf(err, "error selecting documents")
	} else if err == nil {
		defer rows.Close()
		for rows.Next() {
			var tmpKey string
			var tmpTitle string
			var tmpDocument json.RawMessage

			if err := rows.Scan(&tmpKey, &tmpTitle, &tmpDocument); err != nil {
				return nil, errors.Wrapf(err, "error scanning documents")
			}

			profile, ok := profileMap[tmpKey]
			if !ok {
				profile = &Profile{
					Key: tmpKey,
					Documents: []Document{
						{
							Title:    &tmpTitle,
							Document: tmpDocument,
						},
					},
				}
			} else {
				profile.Documents = append(profile.Documents, Document{
					Title:    &tmpTitle,
					Document: tmpDocument,
				})
			}
		}
	}
	// part map to slice
	ret := make([]Profile, 0, len(profileMap))
	for _, value := range profileMap {
		ret = append(ret, *value)
	}

	return ret, nil
}
