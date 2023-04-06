package processor

import (
	"database/sql"
	"encoding/hex"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"wrs/tkdb/goose/packages/archive/tree"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

func NewArchiveProcessor(Tx *sql.Tx, visitArchive func(archive *tree.Archive) error, visitFile func(file *tree.File) error) (*ArchiveProcessor, error) {
	processor := new(ArchiveProcessor)
	processor.Reset()
	processor.Tx = Tx
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
func InitArchive(db *sql.Tx, archiveID int64) (*tree.Archive, error) {
	ret := new(tree.Archive)

	var name, sha1String, sha256String, md5String sql.NullString
	var size sql.NullInt64
	if err := db.QueryRow(`SELECT name, size, checksum_sha1, checksum_sha256, checksum_md5
	FROM archive_table WHERE id=$1`,
		archiveID).Scan(&name, &size, &sha1String, &sha256String, &md5String); err != nil {
		return nil, err
	}

	ret.Name = name.String

	if sha1String.Valid {
		sha1, err := hex.DecodeString(sha1String.String)
		if err != nil {
			return nil, errors.Wrapf(err, "error decoding sha1: %s", sha1String)
		}

		copy(ret.Sha1[:], sha1)
	}
	if md5String.Valid {
		md5, err := hex.DecodeString(md5String.String)
		if err != nil {
			return nil, errors.Wrapf(err, "error decoding md5: %s", md5String.String)
		}
		copy(ret.Md5[:], md5)
	}
	if sha256String.Valid {
		sha256, err := hex.DecodeString(sha256String.String)
		if err != nil {
			return nil, errors.Wrapf(err, "error decoding sha256: %s", sha256String.String)
		}
		copy(ret.Sha256[:], sha256)
	}
	ret.Size = size.Int64

	return ret, nil
}

func upsertSlice[E any](dst []E, element E) []E {
	if dst == nil {
		dst = make([]E, 0)
	}

	return append(dst, element)
}

// trimPath is used to remove the first directory in the path
// This directory is specific to our extraction process, and is not a directory originally of the archive
func trimPath(path string) string {
	if path == "" {
		return ""
	}

	if index := strings.Index(path, "/"); index == -1 {
		return path
	}

	ret := filepath.Join(strings.Split(path, "/")[1:]...)

	return ret
}

func (processor *ArchiveProcessor) ProcessArchive(archiveID int64, archive *tree.Archive) (*tree.Archive, error) {
	if archive == nil {
		var err error
		archive, err = InitArchive(processor.Tx, archiveID)
		if err != nil {
			return archive, err

		}
	}

	// archive already processed
	if a, ok := processor.ArchiveMap[archive.Sha256]; ok {
		return a, nil
	}

	var fileCollectionID sql.NullInt64
	if err := processor.Tx.QueryRow(`SELECT file_collection_id FROM archive_table WHERE id=$1`,
		archiveID).Scan(&fileCollectionID); err != nil {
		return nil, errors.Wrapf(err, "error selecting file_collection_id of archive %d", archiveID)
	} else if !fileCollectionID.Valid {
		return nil, errors.New(fmt.Sprintf("archive %d does not have a file collcetion", archiveID))
	}

	// get data
	if err := processor.Tx.QueryRow(`SELECT le.expression, fc.license_rationale, fc.license_notice
	FROM file_collection fc
	LEFT JOIN license_expression le ON le.id=fc.license_id
	WHERE fc.id=$1`, fileCollectionID).Scan(&archive.License, &archive.LicenseRationale, &archive.LicenseNotice); err != nil {
		return nil, errors.Wrapf(err, "error selecting file_collection data")
	}

	if processor.VisitArchive != nil {
		if err := processor.VisitArchive(archive); err != nil {
			return archive, err
		}
	}

	// process files
	rows, err := processor.Tx.Query(`SELECT f.checksum_sha1, f.checksum_sha256, f.checksum_md5, f.size,
	fbc.path
	FROM file_table f
	INNER JOIN file_belongs_collection fbc ON fbc.file_id=f.id
	WHERE fbc.file_collection_id=$1`,
		fileCollectionID.Int64)
	if err != nil {
		return nil, errors.Wrapf(err, "error selecting files for %d", fileCollectionID.Int64)
	}
	defer rows.Close()

	for rows.Next() {
		tmpFile := new(tree.File)

		var sha1String string
		var sha256String, md5String sql.NullString
		var size sql.NullInt64
		var path string
		if err := rows.Scan(&sha1String, &sha256String, &md5String, &size, &path); err != nil {
			return nil, errors.Wrapf(err, "error scanning file of %d", fileCollectionID.Int64)
		}

		if !sha256String.Valid {
			return nil, ErrSha256
		}

		sha256, err := hex.DecodeString(sha256String.String)
		if err != nil {
			return nil, errors.Wrapf(err, "error decoding sha256: %s", sha256String)
		}

		copy(tmpFile.Sha256[:], sha256)
		if mapFile, ok := processor.FileMap[tmpFile.Sha256]; ok {
			tmpFile = mapFile
		} else {
			sha1, err := hex.DecodeString(sha1String)
			if err != nil {
				return nil, errors.Wrapf(err, "error decoding sha1: %s", sha1String)
			}
			copy(tmpFile.Sha1[:], sha1)

			if md5String.Valid {
				md5, err := hex.DecodeString(md5String.String)
				if err != nil {
					return nil, errors.Wrapf(err, "error decoding md5: %s", md5String.String)
				}
				copy(tmpFile.Md5[:], md5)
			}

			if size.Valid {
				tmpFile.Size = size.Int64
			}

			processor.FileMap[tmpFile.Sha256] = tmpFile
		}

		archive.AddFile(path, tmpFile)
	}
	rows.Close()

	// process sub-archives
	fileCollectionContains := make([]struct {
		ChildID int64
		Path    string
	}, 0)
	rows, err = processor.Tx.Query(`SELECT child_id, path 
	FROM file_collection_contains
	WHERE parent_id=$1`, fileCollectionID.Int64)
	if err != nil {
		return nil, errors.Wrapf(err, "error selecting sub-collections for %d", fileCollectionID.Int64)
	}
	defer rows.Close()

	for rows.Next() {
		var tmp struct {
			ChildID int64
			Path    string
		}

		if err := rows.Scan(&tmp.ChildID, &tmp.Path); err != nil {
			return nil, errors.Wrapf(err, "error scanning sub-collections for %d", fileCollectionID.Int64)
		}

		fileCollectionContains = append(fileCollectionContains, tmp)
	}
	rows.Close()
	log.Debug().Interface("fileCollectionContains", fileCollectionContains).Msg("Processing file_collection_contains")

	for _, v := range fileCollectionContains {
		needle := filepath.Base(v.Path)
		var archiveID sql.NullInt64
		otherArchives := make([]struct {
			ID   int64
			Name string
		}, 0)

		// try to find matching archive
		rows, err := processor.Tx.Query(`SELECT name, id FROM archive_table WHERE file_collection_id=$1`,
			v.ChildID)
		if err != nil {
			return nil, errors.Wrapf(err, "error selecting archives for child collection %d", v.ChildID)
		}
		defer rows.Close()

		for rows.Next() {
			var tmpID int64
			var tmpName sql.NullString

			if err := rows.Scan(&tmpName, &tmpID); err != nil {
				return nil, errors.Wrapf(err, "error scanning archives for child collection %d", v.ChildID)
			}

			if archiveID.Valid { // already found a matching arhive
				otherArchives = append(otherArchives, struct {
					ID   int64
					Name string
				}{
					ID:   tmpID,
					Name: tmpName.String,
				})
			} else if tmpName.String == needle {
				archiveID.Int64 = tmpID
				archiveID.Valid = true
			} else {
				otherArchives = append(otherArchives, struct {
					ID   int64
					Name string
				}{
					ID:   tmpID,
					Name: tmpName.String,
				})
			}
		}
		rows.Close()

		if !archiveID.Valid && len(otherArchives) > 0 {
			archiveID.Int64 = otherArchives[0].ID
			archiveID.Valid = true
		}

		var subNode tree.Node
		if archiveID.Valid {
			// process sub-archive
			subNode, err = processor.ProcessArchive(archiveID.Int64, nil)
			if err != nil {
				return nil, err
			}
		} else {
			// process sub-collection
			subNode, err = processor.ProcessCollection(v.ChildID)
			if err != nil {
				return nil, err
			}
		}

		archive.AddNode(v.Path, subNode)
	}

	return archive, nil
}

func (processor *ArchiveProcessor) ProcessCollection(fileCollectionID int64) (*tree.FileCollection, error) {
	ret := new(tree.FileCollection)

	// get data
	if err := processor.Tx.QueryRow(`SELECT le.expression, fc.license_rationale, fc.license_notice
	FROM file_collection fc
	LEFT JOIN license_expression le ON le.id=fc.license_id
	WHERE fc.id=$1`, fileCollectionID).Scan(&ret.License, &ret.LicenseRationale, &ret.LicenseNotice); err != nil {
		return nil, errors.Wrapf(err, "error selecting file_collection data")
	}

	// process files
	rows, err := processor.Tx.Query(`SELECT f.checksum_sha1, f.checksum_sha256, f.checksum_md5, f.size,
	fbc.path
	FROM file_table f
	INNER JOIN file_belongs_collection fbc ON fbc.file_id=f.id
	WHERE fbc.file_collection_id=$1`,
		fileCollectionID)
	if err != nil {
		return nil, errors.Wrapf(err, "error selecting files for %d", fileCollectionID)
	}
	defer rows.Close()

	for rows.Next() {
		tmpFile := new(tree.File)

		var sha1String string
		var sha256String, md5String sql.NullString
		var size sql.NullInt64
		var path string
		if err := rows.Scan(&sha1String, &sha256String, &md5String, &size, &path); err != nil {
			return nil, errors.Wrapf(err, "error scanning file of %d", fileCollectionID)
		}

		if !sha256String.Valid {
			return nil, ErrSha256
		}

		sha256, err := hex.DecodeString(sha256String.String)
		if err != nil {
			return nil, errors.Wrapf(err, "error decoding sha256: %s", sha256String)
		}

		copy(tmpFile.Sha256[:], sha256)
		if mapFile, ok := processor.FileMap[tmpFile.Sha256]; ok {
			tmpFile = mapFile
		} else {
			sha1, err := hex.DecodeString(sha1String)
			if err != nil {
				return nil, errors.Wrapf(err, "error decoding sha1: %s", sha1String)
			}
			copy(tmpFile.Sha1[:], sha1)

			if md5String.Valid {
				md5, err := hex.DecodeString(md5String.String)
				if err != nil {
					return nil, errors.Wrapf(err, "error decoding md5: %s", md5String.String)
				}
				copy(tmpFile.Md5[:], md5)
			}

			if size.Valid {
				tmpFile.Size = size.Int64
			}

			processor.FileMap[tmpFile.Sha256] = tmpFile
		}

		ret.AddFile(path, tmpFile)
	}
	rows.Close()

	// process sub-archives
	fileCollectionContains := make([]struct {
		ChildID int64
		Path    string
	}, 0)
	rows, err = processor.Tx.Query(`SELECT child_id, path 
	FROM file_collection_contains
	WHERE parent_id=$1`, fileCollectionID)
	if err != nil {
		return nil, errors.Wrapf(err, "error selecting sub-collections for %d", fileCollectionID)
	}
	defer rows.Close()

	for rows.Next() {
		var tmp struct {
			ChildID int64
			Path    string
		}

		if err := rows.Scan(&tmp.ChildID, &tmp.Path); err != nil {
			return nil, errors.Wrapf(err, "error scanning sub-collections for %d", fileCollectionID)
		}

		fileCollectionContains = append(fileCollectionContains, tmp)
	}
	rows.Close()

	for _, v := range fileCollectionContains {
		needle := filepath.Base(v.Path)
		var archiveID sql.NullInt64
		otherArchives := make([]struct {
			ID   int64
			Name string
		}, 0)

		// try to find matching archive
		rows, err := processor.Tx.Query(`SELECT name, id FROM archive_table WHERE file_collection_id=$1`,
			v.ChildID)
		if err != nil {
			return nil, errors.Wrapf(err, "error selecting archives for child collection %d", v.ChildID)
		}
		defer rows.Close()

		for rows.Next() {
			var tmpID int64
			var tmpName sql.NullString

			if err := rows.Scan(&tmpName, &tmpID); err != nil {
				return nil, errors.Wrapf(err, "error scanning archives for child collection %d", v.ChildID)
			}

			if archiveID.Valid { // already found a matching arhive
				otherArchives = append(otherArchives, struct {
					ID   int64
					Name string
				}{
					ID:   tmpID,
					Name: tmpName.String,
				})
			} else if tmpName.String == needle {
				archiveID.Int64 = tmpID
				archiveID.Valid = true
			} else {
				otherArchives = append(otherArchives, struct {
					ID   int64
					Name string
				}{
					ID:   tmpID,
					Name: tmpName.String,
				})
			}
		}
		rows.Close()

		if !archiveID.Valid && len(otherArchives) > 0 {
			archiveID.Int64 = otherArchives[0].ID
			archiveID.Valid = true
		}

		var subNode tree.Node
		if archiveID.Valid {
			// process sub-archive
			subNode, err = processor.ProcessArchive(archiveID.Int64, nil)
			if err != nil {
				return nil, err
			}
		} else {
			// process sub-collection
			subNode, err = processor.ProcessCollection(v.ChildID)
			if err != nil {
				return nil, err
			}
			subNode.SetName(needle)
		}

		ret.AddNode(v.Path, subNode)
	}

	return ret, nil
}

func InitFile(Tx *sql.Tx, fileID int64) (*tree.File, error) {
	ret := new(tree.File)

	var sha1String string
	var sha256String, md5String sql.NullString
	var size sql.NullInt64
	if err := Tx.QueryRow(`SELECT checksum_sha1, checksum_sha256, checksum_md5, size
	FROM file_table
	WHERE id=$1`, fileID).Scan(&sha1String, &sha256String, &md5String, &size); err != nil {
		return nil, errors.Wrapf(err, "error selecting file %d", fileID)
	}

	sha1, err := hex.DecodeString(sha1String)
	if err != nil {
		return nil, errors.Wrapf(err, "error decoding sha1: %s", sha1String)
	}
	copy(ret.Sha1[:], sha1)

	if sha256String.Valid {
		sha256, err := hex.DecodeString(sha256String.String)
		if err != nil {
			return nil, errors.Wrapf(err, "error decoding sha256: %s", sha256String)
		}
		copy(ret.Sha256[:], sha256)
	}

	if md5String.Valid {
		md5, err := hex.DecodeString(md5String.String)
		if err != nil {
			return nil, errors.Wrapf(err, "error decoding md5: %s", md5String.String)
		}
		copy(ret.Md5[:], md5)
	}

	if size.Valid {
		ret.Size = size.Int64
	}

	return ret, nil
}

func IsSymLink(info fs.FileInfo) bool {
	if info.Mode()&os.ModeSymlink > 0 {
		return true
	}

	return false
}

// func (process *ArchiveProcessor) processFile(archive *tree.Archive, path string, info fs.FileInfo) error {
// 	if IsSymLink(info) {
// 		log.Debug().Str("path", path).Interface("info", info).Msg("You shouldn't be processing this symlink")
// 	}
// 	newFile, err := InitFile(process.Tx, )
// 	if err != nil {
// 		return err
// 	}

// 	file, ok := process.FileMap[newFile.Sha256]
// 	if !ok {
// 		process.FileMap[newFile.Sha256] = newFile
// 		file = newFile
// 		if process.VisitFile != nil {
// 			if err := process.VisitFile(path, newFile); err != nil {
// 				return err
// 			}
// 		}
// 	}

// 	archive.Files = upsertSlice[tree.SubFile](archive.Files, tree.SubFile{
// 		Path: trimPath(path),
// 		File: file,
// 	})

// 	return nil
// }
