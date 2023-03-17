package processor

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"wrs/tk/packages/core/archive/tree"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"gitlab.devstar.cloud/ip-systems/extract.git"
)

func NewArchiveProcessor(visitArchive func(archivePath string, archive *tree.Archive) error, visitFile func(filePath string, file *tree.File) error) (*ArchiveProcessor, error) {
	processor := new(ArchiveProcessor)
	processor.Reset()
	processor.VisitArchive = visitArchive
	processor.VisitFile = visitFile

	// log.Debug().Str("rootArchivePath", rootArchive).Msg("Initializing Archive")
	// root, err := InitArchive(rootArchive)
	// if err != nil {
	// 	return processor, err
	// }
	// log.Debug().Str("rootArchivePath", rootArchive).Interface("rootArchive", root).Msg("Initialized Archive")

	// if err := processor.extractArchive(rootArchive, root); err != nil {
	// 	return processor, errors.Wrapf(err, "error extracting root archive")
	// }
	// log.Debug().Str("rootArchivePath", rootArchive).Str("extracted", *root.extracted).Msg("Extracted Archive")

	// root, err = processor.ProcessArchive(rootArchive, root)
	// if err != nil {
	// 	return processor, err
	// }
	// log.Debug().Str("rootArchivePath", rootArchive).Interface("rootArchive", root).Msg("Archive Processed")

	// processor.RootArchive = root

	return processor, nil
}

func (ap *ArchiveProcessor) Reset() {
	ap.ArchiveMap = make(ArchiveMap)
	ap.FileMap = make(FileMap)
}

// Init archive fills in and returns the identifying information on an archive found at archivePath
// The archive itself still needs to be extracted and cataloged to Archive.Files and Archive.Archives
func InitArchive(archivePath string) (*tree.Archive, error) {
	ret := new(tree.Archive)
	ret.TmpPath = &archivePath

	// Chop name out of file path
	ret.Name = filepath.Base(archivePath)

	// Open archive file for further processing
	f, err := os.Open(archivePath)
	if err != nil {
		return ret, errors.Wrapf(err, "error opening %s", archivePath)
	}
	defer f.Close()

	stat, err := f.Stat()
	if err != nil {
		return ret, errors.Wrapf(err, "error statting %s", archivePath)
	}

	// Get size
	ret.Size = stat.Size()

	// Hash archive
	hasherSha256 := sha256.New()
	hasherMd5 := md5.New()
	hasherSha1 := sha1.New()
	hasher := io.MultiWriter(hasherSha256, hasherMd5, hasherSha1)

	if _, err := io.Copy(hasher, f); err != nil {
		return ret, errors.Wrapf(err, "error hashing archive %s", archivePath)
	}

	copy(ret.Sha256[:], hasherSha256.Sum(nil))
	copy(ret.Md5[:], hasherMd5.Sum(nil))
	copy(ret.Sha1[:], hasherSha1.Sum(nil))

	return ret, nil
}

func upsertSlice[E any](dst []E, element E) []E {
	if dst == nil {
		dst = make([]E, 0)
	}

	return append(dst, element)
}

func (processor *ArchiveProcessor) ProcessArchive(archivePath string, archive *tree.Archive) (*tree.Archive, error) {
	if archive == nil {
		var err error
		archive, err = InitArchive(archivePath)
		if err != nil {
			return archive, err

		}
	}

	if err := processor.extractArchive(archivePath, archive); err != nil {
		return archive, err
	}
	defer archive.Close()
	log.Debug().Str("extracted", *archive.Extracted).Msg("extracted archive")

	if processor.VisitArchive != nil {
		if err := processor.VisitArchive(archivePath, archive); err != nil { // TODO upload archive
			return archive, err
		}
	}

	if err := filepath.Walk(*archive.Extracted, func(path string, info fs.FileInfo, err error) error {
		// log.Debug().Str("path", path).Msg("Walked to path")
		if info.IsDir() { // only want to process files, so just skip this entry
			return nil
		} else if !info.Mode().IsRegular() { //skip irregular files
			return nil
		}

		if rec := extract.IsExtractable(path); rec != 1.0 { // path does not look extractable
			// process as normal file
			return processor.processFile(archive, path, info)
		}

		// else process archive

		newArchive, err := InitArchive(path)
		if err != nil {
			return err
		}

		sub, ok := processor.ArchiveMap[newArchive.Sha256]
		if !ok {
			processor.ArchiveMap[newArchive.Sha256] = newArchive
			sub = newArchive
		}

		if err := processor.extractArchive(path, sub); err != nil {
			if _, ok := err.(ErrExtract); ok {
				// log.Debug().Err(err).Str("path", path).Interface("sub", sub).Msg("Error extracting sub-archive, so treating as file")
				return processor.processFile(archive, path, info) // error extracting, so process as file
			}

			// return unexpected error
			return err
		}

		log.Debug().Interface("sub", sub).Str("path", path).Msg("processing sub-archive")
		sub, err = processor.ProcessArchive(path, sub)
		if err != nil {
			return err
		}

		archive.Archives = upsertSlice[tree.SubArchive](archive.Archives, tree.SubArchive{
			Path:    path, // TODO trim os path context
			Archive: sub,
		})

		return nil
	}); err != nil {
		return archive, err
	}

	return archive, nil
}

// ErrExtract is an error returned when the actual extraction step fails
// not returned if an unrelated error, such as failing to read the file at all, occurs
type ErrExtract struct {
	error
}

// extractArchive extracts the given archive, and sets the archives extracted field
// if no error is returned, it can be assumed archive.extracted is not nil
func (process *ArchiveProcessor) extractArchive(path string, archive *tree.Archive) error {
	extractor, err := extract.New(path, archive.Name)
	if err != nil {
		return err
	}
	if err := extractor.Enclose(); err != nil {
		return err
	}

	extracted, err := extractor.Extract()
	if err != nil {
		return ErrExtract{err}
	}

	archive.Extracted = &extracted

	return nil
}

func InitFile(filePath string) (*tree.File, error) {
	ret := new(tree.File)

	// Open file for further processing
	f, err := os.Open(filePath)
	if err != nil {
		return ret, errors.Wrapf(err, "error opening %s", filePath)
	}
	defer f.Close()

	stat, err := f.Stat()
	if err != nil {
		return ret, errors.Wrapf(err, "error statting %s", filePath)
	}

	// Get size
	ret.Size = stat.Size()

	// Hash archive
	hasherSha256 := sha256.New()
	hasherMd5 := md5.New()
	hasherSha1 := sha1.New()
	hasher := io.MultiWriter(hasherSha256, hasherMd5, hasherSha1)

	if _, err := io.Copy(hasher, f); err != nil {
		return ret, errors.Wrapf(err, "error hashing file %s", filePath)
	}

	copy(ret.Sha256[:], hasherSha256.Sum(nil))
	copy(ret.Md5[:], hasherMd5.Sum(nil))
	copy(ret.Sha1[:], hasherSha1.Sum(nil))

	return ret, nil
}

func IsSymLink(info fs.FileInfo) bool {
	if info.Mode()&os.ModeSymlink > 0 {
		return true
	}

	return false
}

func (process *ArchiveProcessor) processFile(archive *tree.Archive, path string, info fs.FileInfo) error {
	if IsSymLink(info) {
		log.Debug().Str("path", path).Interface("info", info).Msg("You shouldn't be processing this symlink")
	}
	newFile, err := InitFile(path)
	if err != nil {
		return err
	}

	file, ok := process.FileMap[newFile.Sha256]
	if !ok {
		process.FileMap[newFile.Sha256] = newFile
		file = newFile
		if process.VisitFile != nil {
			if err := process.VisitFile(path, newFile); err != nil { // TODO upload file
				return err
			}
		}
	}

	archive.Files = upsertSlice[tree.SubFile](archive.Files, tree.SubFile{
		Path: path, // TODO trim
		File: file,
	})

	return nil
}
