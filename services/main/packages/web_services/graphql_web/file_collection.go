package graphql_web

import (
	"time"

	"wrs/tk/packages/core/archive"
	"wrs/tk/packages/core/file_collection"

	"github.com/graphql-go/graphql"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type FileCollection struct {
	ID                int64     `json:"id"`
	InsertDate        time.Time `json:"insert_date"`
	GroupContainerID  int       `json:"group_container_id"`
	Extracted         bool      `json:"flag_extract"`
	LicenseExtracted  bool      `json:"flag_license_extract"`
	LicenseID         int64     `json:"license_id"`
	LicenseRationale  string    `json:"license_rationale"`
	AnalystID         int64     `json:"analyst_id"`
	LicenseExpression string    `json:"license_expression"`
	LicenseNotice     string    `json:"license_notice"`
	Copyright         string    `json:"copyright"`
	FVCOne            []byte    `json:"verification_code_one"`
	FVCTwo            []byte    `json:"verification_code_two"`
}

func toFileCollection(fc *file_collection.FileCollection) FileCollection {
	return FileCollection{
		ID:                fc.FileCollectionID,
		InsertDate:        fc.InsertDate,
		GroupContainerID:  int(fc.GroupID.Int64),
		Extracted:         fc.Extracted,
		LicenseExtracted:  fc.LicenseExtracted,
		LicenseID:         fc.LicenseID.Int64,
		LicenseRationale:  fc.LicenseRationale.String,
		AnalystID:         fc.AnalystID.Int64,
		LicenseExpression: fc.LicenseExpression.String,
		LicenseNotice:     fc.LicenseNotice.String,
		Copyright:         fc.Copyright.String,
		FVCOne:            fc.VerificationCodeOne,
		FVCTwo:            fc.VerificationCodeTwo,
	}
}

var fileCollectionType = graphql.NewObject(
	graphql.ObjectConfig{
		Name: "FileCollection",
		Fields: graphql.Fields{
			"id": &graphql.Field{
				Type: graphql.Int,
			},
			"insert_date": &graphql.Field{
				Type: graphql.DateTime,
			},
			"group_container_id": &graphql.Field{
				Type: graphql.Int,
			},
			"flag_extract": &graphql.Field{
				Type: graphql.Boolean,
			},
			"flag_license_extract": &graphql.Field{
				Type: graphql.Boolean,
			},
			"license_id": &graphql.Field{
				Type: graphql.Int,
			},
			"license_rationale": &graphql.Field{
				Type: graphql.String,
			},
			"analyst_id": &graphql.Field{
				Type: graphql.Int,
			},
			"license_expression": &graphql.Field{
				Type: graphql.String,
			},
			"license_notice": &graphql.Field{
				Type: graphql.String,
			},
			"copyright": &graphql.Field{
				Type: graphql.String,
			},
			"verification_code_one": &graphql.Field{
				Type: graphql.String,
			},
			"verification_code_two": &graphql.Field{
				Type: graphql.String,
			},
		},
	},
)

func resolveFileCollectionQuery(p graphql.ResolveParams) (interface{}, error) {
	logger := log.With().Str(zerolog.CallerFieldName, "resolveFilCollectionQuery").Interface("args", p.Args).Logger()
	logger.Info().Msg("resolving file_collection query")

	fileCollectionController, err := file_collection.GetFileCollectionController(p.Context)
	if err != nil {
		return nil, err
	}
	archiveController, err := archive.GetArchiveController(p.Context)
	if err != nil {
		return nil, err
	}

	if id, ok := p.Args["id"].(int); ok {
		logger.Debug().Msg("resolving file_collection with id")

		fc, err := fileCollectionController.GetByID(int64(id))
		if err != nil {
			return nil, err
		}

		return toFileCollection(fc), nil
	}
	if sha256, ok := p.Args["sha256"].(string); ok {
		logger.Debug().Msg("resolving archive with sha256")
		a, err := archiveController.GetBySha256(sha256)
		if err != nil {
			return nil, err
		}

		fc, err := fileCollectionController.GetByID(a.FileCollectionID.Int64)
		if err != nil {
			return nil, err
		}

		return toFileCollection(fc), nil
	}
	if sha1, ok := p.Args["sha1"].(string); ok {
		logger.Debug().Msg("resolving archive with sha1")

		a, err := archiveController.GetBySha1(sha1)
		if err != nil {
			return nil, err
		}

		fc, err := fileCollectionController.GetByID(a.FileCollectionID.Int64)
		if err != nil {
			return nil, err
		}

		return toFileCollection(fc), nil
	}
	if name, ok := p.Args["name"].(string); ok {
		logger.Debug().Msg("resolving archive with name")

		a, err := archiveController.GetByName(name)
		if err != nil {
			return nil, err
		}

		fc, err := fileCollectionController.GetByID(a.FileCollectionID.Int64)
		if err != nil {
			return nil, err
		}

		return toFileCollection(fc), nil
	}

	return nil, nil
}
