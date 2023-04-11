// Copyright (c) 2020 Wind River Systems, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at:
//       http://www.apache.org/licenses/LICENSE-2.0
// Unless required by applicable law or agreed to in writing, software  distributed
// under the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES
// OR CONDITIONS OF ANY KIND, either express or implied.

package partlist

import (
	"database/sql"
	"wrs/tk/packages/core/part"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

type PartList struct {
	ID        int64         `json:"id"`
	Name      string        `json:"name"`
	Parent_ID sql.NullInt64 `json:"parent_id"`
}

type PartListController struct {
	DB *sqlx.DB
}

func (controller PartListController) AddParts(id int64, parts []*uuid.UUID) (*PartList, error) {
	for _, v := range parts {
		result, err := controller.DB.Exec("INSERT INTO partlist_has_part(partlist_id, part_id) VALUES ($1, $2)", id, *v)
		if err != nil {
			return nil, err
		}
		if count, _ := result.RowsAffected(); count < 1 {
			return nil, errors.New("unable to insert partlist, no rows affected")
		}
	}
	var ret PartList
	if err := controller.DB.QueryRowx("SELECT * FROM partlist WHERE id=$1", id).StructScan(&ret); err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &ret, nil
}

func (controller PartListController) AddPartList(name string) (*PartList, error) {
	var ret PartList
	result, err := controller.DB.Exec("INSERT INTO partlist(name) VALUES ($1)", name)
	if err != nil {
		return nil, err
	}
	if count, _ := result.RowsAffected(); count < 1 {
		return nil, errors.New("unable to insert partlist, no rows affected")
	}
	if err := controller.DB.QueryRowx("SELECT * FROM partlist WHERE name=$1", name).StructScan(&ret); err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &ret, nil
}

func (controller PartListController) AddPartListWithParent(name string, parentID int64) (*PartList, error) {
	var ret PartList
	result, err := controller.DB.Exec("INSERT INTO partlist(name, parent_id) VALUES ($1, $2)", name, parentID)
	if err != nil {
		return nil, err
	}
	if count, _ := result.RowsAffected(); count < 1 {
		return nil, errors.New("unable to insert partlist, no rows affected")
	}
	if err := controller.DB.QueryRowx("SELECT * FROM partlist WHERE name=$1", name).StructScan(&ret); err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &ret, nil
}

func (controller PartListController) DeletePartList(id int64) (*PartList, error) {
	var ret PartList
	if err := controller.DB.QueryRowx("SELECT * FROM partlist WHERE id=$1", id).StructScan(&ret); err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, err
	}
	result, err := controller.DB.Exec("DELETE FROM partlist WHERE id=$1", id)
	if err != nil {
		return nil, err
	}
	if count, _ := result.RowsAffected(); count < 1 {
		return nil, errors.New("unable to delete partlist, no rows affected")
	}
	return &ret, nil
}

func (controller PartListController) DeletePartFromList(list_id int64, part_id uuid.UUID) (*PartList, error) {
	var ret PartList
	if err := controller.DB.QueryRowx("SELECT * FROM partlist WHERE id=$1", list_id).StructScan(&ret); err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, err
	}
	result, err := controller.DB.Exec("DELETE FROM partlist_has_part WHERE partlist_id=$1 AND part_id=$2", list_id, part_id)
	if err != nil {
		return nil, err
	}
	if count, _ := result.RowsAffected(); count < 1 {
		return nil, errors.New("unable to delete partlist, no rows affected")
	}
	return &ret, nil
}

func (controller PartListController) GetByID(partlistID int64) (*PartList, error) {
	var ret PartList
	if err := controller.DB.QueryRowx("SELECT * FROM partlist WHERE id=$1",
		partlistID).StructScan(&ret); err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}

		return nil, err
	}

	return &ret, nil
}

func (controller PartListController) GetByName(name string) (*PartList, error) {
	var ret PartList
	if err := controller.DB.QueryRowx("SELECT * FROM partlist WHERE name=$1",
		name).StructScan(&ret); err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}

		return nil, err
	}

	return &ret, nil
}

func (controller PartListController) GetByParentID(parentID int64) ([]PartList, error) {
	ret := make([]PartList, 0)

	var rows *sqlx.Rows
	var err error

	if parentID == 0 {
		rows, err = controller.DB.Queryx("SELECT * FROM partlist WHERE parent_id IS NULL")
	} else {
		rows, err = controller.DB.Queryx("SELECT * FROM partlist WHERE parent_id=$1",
			parentID)
	}
	if err != nil {
		return nil, errors.Wrapf(err, "error selecting partlists by parent_id")
	}
	defer rows.Close()

	for rows.Next() {
		var pl PartList
		if err := rows.StructScan(&pl); err != nil {
			return nil, errors.Wrapf(err, "error scanning partlists by parent_id")
		}

		ret = append(ret, pl)
	}
	rows.Close()

	return ret, nil
}

type PartListHasPart struct {
	Partlist_id int64     `db:"partlist_id"`
	Part_id     uuid.UUID `db:"part_id"`
}

func (controller PartListController) GetParts(id int64) ([]*part.Part, error) {
	ret := make([]*part.Part, 0)
	plHasParts := make([]PartListHasPart, 0)

	rows, err := controller.DB.Queryx("SELECT * FROM partlist_has_part WHERE partlist_id=$1", id)
	if err != nil {
		return nil, errors.Wrapf(err, "error selecting partlists by parent_id")
	}
	defer rows.Close()

	for rows.Next() {
		var plHasPart PartListHasPart
		if err := rows.StructScan(&plHasPart); err != nil {
			return nil, errors.Wrapf(err, "error scanning relation by partlist_id")
		}

		plHasParts = append(plHasParts, plHasPart)
	}
	rows.Close()

	for _, v := range plHasParts {
		var part part.Part
		if err := controller.DB.QueryRowx("SELECT * FROM part WHERE part_id=$1", v.Part_id).StructScan(&part); err != nil {
			return nil, err
		}
		ret = append(ret, &part)
	}

	return ret, nil
}
