package tree

import (
	"os"
	"path/filepath"

	"golang.org/x/text/runes"
)

type Sha256 [32]byte

type File struct {
	Sha256 Sha256
	Size   int64
	Md5    [16]byte
	Sha1   [20]byte
}

type SubFile struct {
	*File
	Path string
}

func (sf SubFile) GetPath() string {
	return runes.ReplaceIllFormed().String(sf.Path)
}

func (sf SubFile) GetName() string {
	return runes.ReplaceIllFormed().String(filepath.Base(sf.Path))
}

type ArchiveIdentifiers struct {
	// Identifying information
	Sha256 Sha256
	Size   int64
	Md5    [16]byte
	Sha1   [20]byte
	// Misc
	Name string
}

func (i ArchiveIdentifiers) GetName() string {
	return runes.ReplaceIllFormed().String(i.Name)
}

type Archive struct {
	ArchiveIdentifiers
	// Relationships
	Files    []SubFile
	Archives []SubArchive

	TmpPath              *string
	Extracted            *string
	FileVerificationCode []byte
	DuplicateArchives    []ArchiveIdentifiers // All archives should be inserted into the database, but the purpose of the trees is actually to turn them into parts, so a separate list of duplicates is required
}

func (a *Archive) Close() error {
	if a.Extracted != nil { // TODO extracting to strange directory
		// log.Debug().Str("extracted", *a.extracted).Msg("Leaving extracted directory for debugging purposes")
		return os.RemoveAll(*a.Extracted)
	}
	if a.TmpPath != nil {
		return os.Remove(*a.TmpPath)
	}

	return nil
}

type SubArchive struct {
	*Archive
	Path string
}
