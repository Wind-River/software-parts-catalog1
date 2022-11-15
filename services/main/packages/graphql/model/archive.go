package model

import (
	"time"
	"wrs/tk/packages/core/archive"
)

type Archive struct {
	ID               int64     `json:"id"`
	FileCollectionID *int64    `json:"file_collection_id,omitempty"`
	Name             string    `json:"name"`
	Path             string    `json:"path"`
	Size             int64     `json:"size"`
	Sha1             string    `json:"sha1"`
	Sha256           string    `json:"sha256"`
	Md5              string    `json:"md5"`
	InsertDate       time.Time `json:"insert_date"`
	ExtractStatus    int       `json:"extract_status"`
}

func ToArchive(a *archive.Archive) Archive {
	ret := Archive{
		ID: a.ArchiveID,
		// FileCollectionID: &a.FileCollectionID.Int64,
		Name:          a.Name.String,
		Path:          a.Path.String,
		Size:          a.Size.Int64,
		Sha1:          a.Sha1.String,
		Sha256:        a.Sha256.String,
		Md5:           a.Md5.String,
		InsertDate:    a.InsertDate,
		ExtractStatus: a.ExtractStatus,
	}

	if a.FileCollectionID.Valid {
		ret.FileCollectionID = &a.FileCollectionID.Int64
	}

	return ret
}
