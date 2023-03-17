package model

import (
	"time"
	"wrs/tk/packages/core/archive"
	"wrs/tk/packages/core/part"
)

type Archive struct {
	Sha256     [32]byte  `json:"sha256"`
	Size       int64     `json:"size"`
	PartID     *part.ID  `json:"part_id"`
	Md5        [16]byte  `json:"md5"`
	Sha1       [20]byte  `json:"sha1"`
	Name       string    `json:"name"`
	InsertDate time.Time `json:"insert_date"`
	// Extracted  bool      `json:"extract_status"`
}

func ToArchive(a *archive.Archive) Archive {
	ret := Archive{
		Sha256:     a.Sha256,
		Size:       a.Size,
		PartID:     a.PartID,
		Md5:        a.Md5,
		Sha1:       a.Sha1,
		InsertDate: a.InsertDate,
		// Extracted?
	}

	if len(a.Aliases) > 0 {
		ret.Name = a.Aliases[0]
	}

	return ret
}
