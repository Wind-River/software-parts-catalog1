// Copyright (c) 2020 Wind River Systems, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at:
//       http://www.apache.org/licenses/LICENSE-2.0
// Unless required by applicable law or agreed to in writing, software  distributed
// under the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES
// OR CONDITIONS OF ANY KIND, either express or implied.

package part

import (
	generic "wrs/tk/packages/generics/graph"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

// PartGraph is meant to cleanly make an file collection -> sub file collection graph
type PartGraph struct {
	ID    string
	Edges []string
	Graph *generic.DirectedGraph[string, string]
}

func NewPartGraph(db *sqlx.DB, partID uuid.UUID) (*PartGraph, error) {
	pgraph := new(PartGraph)
	pgraph.ID = partID.String()
	pgraph.Edges = make([]string, 0)
	pgraph.Graph = generic.NewDirectedGraph[string, string]()

	// Add root's sub-parts to edges
	rows, err := db.Queryx("SELECT child_id FROM part_has_part WHERE parent_id=$1", partID)
	if err != nil {
		return nil, errors.Wrapf(err, "error selecting part's direct children")
	}
	defer rows.Close()

	for rows.Next() {
		var tmp uuid.UUID
		if err := rows.Scan(&tmp); err != nil {
			return nil, errors.Wrapf(err, "error scanning part's direct children")
		}

		pgraph.Edges = append(pgraph.Edges, tmp.String())
		pgraph.Graph.Insert(tmp.String(), tmp.String())
	}
	rows.Close()

	// Traverse edges and add their edges to the graph
	if len(pgraph.Edges) > 0 {
		if err := pgraph.Graph.TraverseUniqueEdges(func(id string) error {
			if id == pgraph.ID { // skip root node
				return nil
			}

			currentNode := pgraph.Graph.Get(id)
			rows, err := db.Queryx("SELECT child_id FROM part_has_part WHERE parent_id=$1", id) // TODO do I need to parse the uuid string for this query?
			if err != nil {
				return errors.Wrapf(err, "error selecting part's childern")
			}
			defer rows.Close()

			for rows.Next() {
				var tmp uuid.UUID
				if err := rows.Scan(&tmp); err != nil {
					return errors.Wrapf(err, "error scanning part's children")
				}

				currentNode.Edges.Add(pgraph.Graph.Insert(tmp.String(), tmp.String()))
			}

			return nil
		}, pgraph.Edges...); err != nil {
			return nil, err
		}
	}

	return pgraph, nil
}

func (fcg PartGraph) TraverseUniqueEdges(visitor func(id string) error) error {
	if len(fcg.Edges) > 0 {
		return fcg.Graph.TraverseUniqueEdges(visitor, fcg.Edges...)
	}

	return nil
}
