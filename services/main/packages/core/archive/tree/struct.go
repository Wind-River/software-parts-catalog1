package tree

import (
	"os"
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

type Archive struct {
	// Identifying information
	Sha256 Sha256
	Size   int64
	Md5    [16]byte
	Sha1   [20]byte
	// Misc
	Name string
	// Relationships
	Files    []SubFile
	Archives []SubArchive

	TmpPath              *string
	Extracted            *string
	FileVerificationCode []byte
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
