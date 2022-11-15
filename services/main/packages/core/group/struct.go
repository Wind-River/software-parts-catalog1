// Copyright (c) 2020 Wind River Systems, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at:
//       http://www.apache.org/licenses/LICENSE-2.0
// Unless required by applicable law or agreed to in writing, software  distributed
// under the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES
// OR CONDITIONS OF ANY KIND, either express or implied.

package group

import (
	"encoding/json"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"gopkg.in/guregu/null.v4"
)

type Container struct {
	ID        int64    `json:"file_collection_id"`
	Names     []string `json:"names"`
	GID       int64    `json:"group_id"`
	GroupName string   `json:"group"`
}

type Group struct {
	GroupID             int64       `db:"id"`
	Name                string      `db:"name"`
	Type                null.String `db:"type"`
	AssociatedLicense   null.String `db:"associatedlicense"`
	AssociatedRationale null.String `db:"associatedrationale"`
	Description         null.String `db:"description"`
	Comments            null.String `db:"comments"`
	InsertDate          time.Time   `db:"insert_date"`
	ParentID            null.Int    `db:"parent_id"`
}

type GroupController struct {
	DB *sqlx.DB
}

func (groups *GroupController) ListContainers(groupID int64) ([]Container, error) {
	query := "WITH RECURSIVE groups AS (" +
		"SELECT g.id, g.parent_id FROM group_container g WHERE g.id=$1 " +
		"UNION SELECT s.id, s.parent_id FROM group_container s " +
		"INNER JOIN groups ON s.parent_id=groups.id" +
		") " +
		"SELECT c.id, c.group_container_id, (SELECT build_group_path(g.id)) as path, ARRAY_TO_JSON(ARRAY(SELECT DISTINCT name FROM archive WHERE archive.file_collection_id=c.id)) as names " +
		/*Artifical Error "SELECT c.id, c.group_container_id, (SELECT build_group_path(g.id)) as path, ARRAY(SELECT DISTINCT name FROM archive WHERE archive.file_collection_id=c.id) as names " + //*/
		"FROM file_collection c " +
		"INNER JOIN group_container g ON g.id=c.group_container_id " +
		"WHERE g.id IN (SELECT id FROM groups)"
	rows, err := groups.DB.Queryx(query, groupID)
	if err != nil {
		return nil, errors.Wrapf(err, "error selecting containers from group %d", groupID)
	}
	defer rows.Close()

	ret := make([]Container, 0)
	for rows.Next() {
		var fileCollectionID int64
		var jsonNames []uint8
		var groupID int64
		var groupName string

		if err = rows.Scan(&fileCollectionID, &groupID, &groupName, &jsonNames); err != nil {
			return nil, errors.Wrapf(err, "error scanning groups")
		}

		var names []string
		if err := json.Unmarshal(jsonNames, &names); err != nil {
			return nil, err
		}

		ret = append(ret, Container{
			ID:        fileCollectionID,
			Names:     names,
			GID:       groupID,
			GroupName: groupName,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, errors.Wrapf(err, "error iterating group rows")
	}

	return ret, nil
}

func (groups *GroupController) ParsePath(path string) (int64, error) { // TODO why is this DB side?
	gid := int64(0)
	err := groups.DB.Get(&gid, "SELECT parse_group_path($1)", path)

	return gid, err
}

func (groups *GroupController) GetByID(groupID int64) (*Group, error) {
	var ret Group
	if err := groups.DB.QueryRowx("SELECT * FROM group_container WHERE id=$1",
		groupID).StructScan(&ret); err != nil {
		return nil, errors.Wrapf(err, "error selecting group")
	}

	return &ret, nil
}

func (groups *GroupController) GetByParentID(parentID int64) ([]Group, error) {
	rows, err := groups.DB.Queryx("SELECT * FROM group_container WHERE parent_id=$1",
		parentID)
	if err != nil {
		return nil, errors.Wrapf(err, "error selecting sub groups")
	}
	defer rows.Close()

	ret := make([]Group, 0)
	for rows.Next() {
		var tmp Group
		if err := rows.StructScan(&tmp); err != nil {
			return ret, errors.Wrapf(err, "error scanning sub groups")
		}

		ret = append(ret, tmp)
	}

	return ret, nil
}

func (groups *GroupController) BuildGroupPath(groupID int64) (string, error) {
	var ret string
	if err := groups.DB.Get(&ret,
		"SELECT build_group_path($1)",
		groupID); err != nil {
		return "", errors.Wrapf(err, "error building group path")
	}

	return ret, nil
}

func (groups *GroupController) CountRelations(groupID int64) (groupCount int, collectionCount int, err error) {
	if err := groups.DB.QueryRowx("SELECT (SELECT COUNT(*) FROM group_container WHERE parent_id=gc.id), (SELECT COUNT(*) FROM file_collection WHERE group_container_id=gc.id) "+
		"FROM group_container gc WHERE id=$1",
		groupID).Scan(&groupCount, &collectionCount); err != nil {
		return groupCount, collectionCount, errors.Wrapf(err, "error counting group relations")
	}

	return
}
