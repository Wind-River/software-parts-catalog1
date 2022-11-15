package graphql_web

import (
	"errors"
	"time"

	"wrs/tk/packages/core/archive"

	"github.com/graphql-go/graphql"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
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

func toArchive(a *archive.Archive) Archive {
	return Archive{
		ID:               a.ArchiveID,
		FileCollectionID: &a.FileCollectionID.Int64,
		Name:             a.Name.String,
		Path:             a.Path.String,
		Size:             a.Size.Int64,
		Sha1:             a.Sha1.String,
		Sha256:           a.Sha256.String,
		Md5:              a.Md5.String,
		InsertDate:       a.InsertDate,
		ExtractStatus:    a.ExtractStatus,
	}
}

var archiveType = graphql.NewObject(
	graphql.ObjectConfig{
		Name: "Archive",
		Fields: graphql.Fields{
			"id": &graphql.Field{
				Type: graphql.Int,
			},
			"file_collection_id": &graphql.Field{
				Type: graphql.Int,
			},
			"name": &graphql.Field{
				Type: graphql.String,
			},
			"path": &graphql.Field{
				Type: graphql.String,
			},
			"size": &graphql.Field{
				Type: graphql.Int,
			},
			"sha1": &graphql.Field{
				Type: graphql.String,
			},
			"sha256": &graphql.Field{
				Type: graphql.String,
			},
			"md5": &graphql.Field{
				Type: graphql.String,
			},
			"insert_date": &graphql.Field{
				Type: graphql.DateTime,
			},
			"extract_status": &graphql.Field{
				Type: graphql.Int,
			},
		},
	},
)

func resolveArchiveQuery(p graphql.ResolveParams) (interface{}, error) {
	logger := log.With().Str(zerolog.CallerFieldName, "resolveArchiveQuery").Interface("args", p.Args).Logger()
	logger.Info().Msg("resolving archive query")

	archiveController, err := archive.GetArchiveController(p.Context)
	if err != nil {
		return nil, err
	}

	if id, ok := p.Args["id"].(int); ok {
		logger.Debug().Msg("resolving archive with id")

		a, err := archiveController.GetByID(int64(id))
		if err != nil {
			return nil, err
		}

		return toArchive(a), nil
	}
	if sha256, ok := p.Args["sha256"].(string); ok {
		logger.Debug().Msg("resolving archive with sha256")
		a, err := archiveController.GetBySha256(sha256)
		if err != nil {
			return nil, err
		}

		return toArchive(a), nil
	}
	if sha1, ok := p.Args["sha1"].(string); ok {
		logger.Debug().Msg("resolving archive with sha1")

		a, err := archiveController.GetBySha1(sha1)
		if err != nil {
			return nil, err
		}

		return toArchive(a), nil
	}
	if name, ok := p.Args["name"].(string); ok {
		logger.Debug().Msg("resolving archive with name")

		a, err := archiveController.GetByName(name)
		if err != nil {
			return nil, err
		}

		return toArchive(a), nil
	}

	return nil, nil
}

func resolveArchiveSearch(p graphql.ResolveParams) (interface{}, error) {
	logger := log.With().Str(zerolog.CallerFieldName, "resolveArchiveSearch").Interface("args", p.Args).Logger()
	logger.Info().Msg("resolving archive search")

	archiveController, err := archive.GetArchiveController(p.Context)
	if err != nil {
		return nil, err
	}

	query, ok := p.Args["query"].(string)
	if !ok {
		return nil, errors.New("missing argument 'query'")
	}

	methodArg := p.Args["method"].(string) // method has a default value
	method := archive.ParseMethod(methodArg)

	distanceResults, err := archiveController.SearchForArchiveAll(query, method)
	if err != nil {
		return nil, err
	}

	archives := make([]Archive, len(distanceResults))
	for i, v := range distanceResults {
		a, err := archiveController.GetByID(v.ArchiveID)
		if err != nil {
			return nil, err
		}

		archives[i] = toArchive(a)
	}

	return archives, nil
}
