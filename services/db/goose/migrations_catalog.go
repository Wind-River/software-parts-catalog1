//go:build catalog

package main

// This imports any catalog migration steps written in golang.
// This is in a separate tagged file from main so that golang migrations from blob are not run when catalog is migrated.
import _ "wrs/tkdb/goose/migrations/catalog"
