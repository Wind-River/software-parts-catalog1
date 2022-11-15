package graphql

//go:generate go run github.com/99designs/gqlgen generate

import (
	"wrs/tk/packages/core/archive"
	"wrs/tk/packages/core/file_collection"
	"wrs/tk/packages/core/license"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

type Resolver struct {
	ArchiveController        *archive.ArchiveController
	FileCollectionController *file_collection.FileCollectionController
	LicenseController        *license.LicenseController
}
