package processor

import "wrs/tk/packages/core/archive/tree"

type ArchiveMap map[tree.Sha256]*tree.Archive
type FileMap map[tree.Sha256]*tree.File

type ArchiveProcessor struct {
	ArchiveMap   ArchiveMap
	FileMap      FileMap
	VisitArchive func(archivePath string, archive *tree.Archive) error
	VisitFile    func(filePath string, file *tree.File) error
}
