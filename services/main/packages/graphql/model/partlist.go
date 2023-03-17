package model

import (
	"wrs/tk/packages/core/partlist"
)

type PartList struct {
	ID        int64
	Name      string
	Parent_ID int64
}

func ToPartList(p *partlist.PartList) PartList {
	ret := PartList{
		ID:        p.ID,
		Name:      p.Name,
		Parent_ID: p.Parent_ID.Int64,
	}

	return ret
}
