package graphql

//go:generate go run github.com/99designs/gqlgen generate

import (
	"wrs/tk/packages/core/archive"
	"wrs/tk/packages/core/license"
	"wrs/tk/packages/core/part"
	"wrs/tk/packages/core/partlist"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

type Resolver struct {
	ArchiveController  *archive.ArchiveController
	PartController     *part.PartController
	LicenseController  *license.LicenseController
	PartListController *partlist.PartListController
}
