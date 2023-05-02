//go:build blob

package main

// This imports any blob migration steps written in golang.
// This is in a separate tagged file from main so that golang migrations from catalog are not run when blob is migrated.
import _ "wrs/tkdb/goose/migrations/blob"
