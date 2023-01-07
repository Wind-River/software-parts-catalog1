package file_collection

import (
	generic "wrs/tk/packages/generics/graph"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

// FileCollectionGraph is meant to cleanly make an file collection -> sub file collection graph
type FileCollectionGraph struct {
	ID    int64
	Edges []int64
	Graph *generic.DirectedGraph[int64, int64]
}

func NewFileCollectionGraph(db *sqlx.DB, fileCollectionID int64) (*FileCollectionGraph, error) {
	fcg := new(FileCollectionGraph)
	fcg.ID = fileCollectionID
	fcg.Edges = make([]int64, 0)
	fcg.Graph = generic.NewDirectedGraph[int64, int64]()

	rows, err := db.Queryx("SELECT child_id FROM file_collection_contains WHERE parent_id=$1", fileCollectionID)
	if err != nil {
		return nil, errors.Wrapf(err, "error selecting file collection's direct children")
	}
	defer rows.Close()

	for rows.Next() {
		var tmp int64
		if err := rows.Scan(&tmp); err != nil {
			return nil, errors.Wrapf(err, "error scanning file collection's direct children")
		}

		fcg.Edges = append(fcg.Edges, tmp)
		fcg.Graph.Insert(tmp, tmp)
	}
	rows.Close()

	if len(fcg.Edges) > 0 {
		if err := fcg.Graph.TraverseUniqueEdges(func(id int64) error {
			if id == fcg.ID { // skip root node
				return nil
			}

			currentNode := fcg.Graph.Get(id)
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

				currentNode.Edges.Add(fcg.Graph.Insert(tmp, tmp))
			}

			return nil
		}, fcg.Edges...); err != nil {
			return nil, err
		}
	}

	return fcg, nil
}

func (fcg FileCollectionGraph) TraverseUniqueEdges(visitor func(id int64) error) error {
	if len(fcg.Edges) > 0 {
		return fcg.Graph.TraverseUniqueEdges(visitor, fcg.Edges...)
	}

	return nil
}
