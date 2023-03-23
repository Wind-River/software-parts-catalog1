// Copyright (c) 2020 Wind River Systems, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at:
//       http://www.apache.org/licenses/LICENSE-2.0
// Unless required by applicable law or agreed to in writing, software  distributed
// under the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES
// OR CONDITIONS OF ANY KIND, either express or implied.

package archive

import (
	generic "wrs/tk/packages/generics/graph"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

// ArchiveGraph is meant to cleanly make an archive -> sub file collection graph
type ArchiveGraph struct {
	ID    int64
	Edges []int64
	Graph *generic.DirectedGraph[int64, int64]
}

// TODO WSTRPG-86; The whole extraction process needs a look
func NewArchiveGraph(db *sqlx.DB, archiveID int64) (*ArchiveGraph, error) {
	ag := new(ArchiveGraph)
	ag.ID = archiveID
	ag.Edges = make([]int64, 0)
	ag.Graph = generic.NewDirectedGraph[int64, int64]()

	rows, err := db.Queryx("SELECT child_id FROM archive_contains WHERE parent_id=$1", archiveID)
	if err != nil {
		return nil, errors.Wrapf(err, "error selecting archive's direct children")
	}
	defer rows.Close()

	for rows.Next() {
		var tmp int64
		if err := rows.Scan(&tmp); err != nil {
			return nil, errors.Wrapf(err, "error scanning archive's direct children")
		}

		ag.Edges = append(ag.Edges, tmp)
		ag.Graph.Insert(tmp, tmp)
	}
	rows.Close()

	if len(ag.Edges) > 0 {
		if err := ag.Graph.TraverseUniqueEdges(func(id int64) error {
			currentNode := ag.Graph.Get(id)
			rows, err := db.Queryx("SELECT child_id FROM file_collection_contains WHERE parent_id=$1", id)
			if err != nil {
				return errors.Wrapf(err, "error selecting file_collection's childern")
			}
			defer rows.Close()

			for rows.Next() {
				var tmp int64
				if err := rows.Scan(&tmp); err != nil {
					return errors.Wrapf(err, "error scanning file_collection's children")
				}

				currentNode.Edges.Add(ag.Graph.Insert(tmp, tmp))
			}

			return nil
		}, ag.Edges...); err != nil {
			return nil, err
		}
	}

	return ag, nil
}

func (ag ArchiveGraph) TraverseUniqueEdges(visitor func(id int64) error) error {
	if len(ag.Edges) > 0 {
		return ag.Graph.TraverseUniqueEdges(visitor, ag.Edges...)
	}

	return nil
}
