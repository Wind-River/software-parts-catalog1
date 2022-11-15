package graphql_web

import (
	"context"
	"net/http"

	"github.com/rs/zerolog/log"

	"github.com/graphql-go/graphql"
	"github.com/graphql-go/handler"
)

var schema graphql.Schema

func init() {
	var err error
	schema, err = graphql.NewSchema(graphql.SchemaConfig{
		Query: graphql.NewObject(graphql.ObjectConfig{
			Name: "PartsCatalog",
			Fields: graphql.Fields{
				"archive": &graphql.Field{
					Type:        archiveType,
					Description: "Lookup Archive",
					Args: graphql.FieldConfigArgument{
						"id": &graphql.ArgumentConfig{
							Type: graphql.Int,
						},
						"sha256": &graphql.ArgumentConfig{
							Type: graphql.String,
						},
						"sha1": &graphql.ArgumentConfig{
							Type: graphql.String,
						},
						"name": &graphql.ArgumentConfig{
							Type: graphql.String,
						},
					},
					Resolve: resolveArchiveQuery,
				},
				"archive_search": &graphql.Field{
					Type:        graphql.NewList(archiveType),
					Description: "Search for Archives based on name",
					Args: graphql.FieldConfigArgument{
						"query": &graphql.ArgumentConfig{
							Type: graphql.String,
						},
						"method": &graphql.ArgumentConfig{
							Type:         graphql.String,
							DefaultValue: "levenshtein",
						},
					},
					Resolve: resolveArchiveSearch,
				},
				"file_collection": &graphql.Field{
					Type:        fileCollectionType,
					Description: "Lookup File Collection",
					Args: graphql.FieldConfigArgument{
						"id": &graphql.ArgumentConfig{
							Type: graphql.Int,
						},
						"sha256": &graphql.ArgumentConfig{
							Type: graphql.String,
						},
						"sha1": &graphql.ArgumentConfig{
							Type: graphql.String,
						},
						"name": &graphql.ArgumentConfig{
							Type: graphql.String,
						},
					},
					Resolve: resolveFileCollectionQuery,
				},
			},
		}),
	})
	if err != nil {
		log.Fatal().Err(err).Msg("error initializing graphql schema")
	}
}

func GraphqlHandler() http.Handler {
	return handler.New(&handler.Config{
		Schema:     &schema,
		Pretty:     true,
		Playground: true,
	})
}

func HandleGraphql(w http.ResponseWriter, r *http.Request) {
	h := handler.New(&handler.Config{
		Schema:     &schema,
		Pretty:     true,
		Playground: true,
	})

	ctx := context.WithValue(r.Context(), "sanity", 42)

	h.ServeHTTP(w, r.WithContext(ctx))
}
