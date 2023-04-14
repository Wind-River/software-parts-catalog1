package processor

import (
	"database/sql"
	"wrs/tkdb/goose/packages/archive/tree"
)

type ArchiveMap map[tree.Sha256]*tree.Archive
type FileMap map[tree.Sha256]*tree.File

type ArchiveProcessor struct {
	Tx           *sql.Tx
	ArchiveMap   ArchiveMap
	FileMap      FileMap
	VisitArchive func(archive *tree.Archive) error
	VisitFile    func(file *tree.File) error
}
