// Copyright (c) 2020 Wind River Systems, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at:
//       http://www.apache.org/licenses/LICENSE-2.0
// Unless required by applicable law or agreed to in writing, software  distributed
// under the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES
// OR CONDITIONS OF ANY KIND, either express or implied.

package archive

import (
	"encoding/hex"
	"os"
	"path/filepath"
	"strings"
	"wrs/tk/packages/blob"
	"wrs/tk/packages/encoding/unicode"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"gitlab.devstar.cloud/ip-systems/extract.git"
	"gitlab.devstar.cloud/ip-systems/extract.git/analysis"
)

// ProcessFileCollection catalogs the files of the given archive that were extracted to parentDirectory, and stores them using blobStorage.
// Handling sub-packages is a recursive process that uses a package graph created here to avoid repeating packages.
// A package verification code is calculated and returned at the end.
func (p *ArchiveController) ProcessFileCollection(archive *Archive, parentDirectory string, blobStorage blob.Storage) (vcodeOne []byte, vcodeTwo []byte, err error) {
	packageGraph := analysis.NewPackageGraph()

	collectionNode, err := packageGraph.InsertHexString(archive.Name.String, archive.Path.String, int(archive.Size.Int64), archive.Sha1.String, archive.Sha256.String)
	if err != nil {
		return nil, nil, err
	}

	return p.processFileCollection(p.DB, packageGraph, collectionNode,
		archive, parentDirectory, blobStorage)
}

// processFileCollection catalogs files found at parentDirectory as children of this archive, and recursively processes any sub-packages.
func (p *ArchiveController) processFileCollection(db *sqlx.DB, packageGraph *analysis.PackageGraph, collectionNode *analysis.PackageNode,
	arch *Archive, parentDirectory string, blobStorage blob.Storage) (vcodeOne []byte, vcodeTwo []byte, err error) {
	// Traverse all files
	dq := make(DirectoryQueue, 0)
	rootInfo, err := os.ReadDir(parentDirectory)
	if err != nil {
		err = errors.Wrapf(err, "error reading directory %s", parentDirectory)
		return nil, nil, err
	}

	// processAsFileFunc loads a file path as a file.
	processAsFileFunc := func(db *sqlx.DB, filePath string) error {
		fileStat, err := os.Lstat(filePath)
		if err != nil {
			err = errors.Wrapf(err, "error stating %s", filePath)
			return err
		}

		f, err := NewFile(filePath)
		if err != nil {
			return err
		}

		// Insert file + file_alias
		var faid int64
		if err = db.QueryRowx("SELECT insert_file($1::TEXT, $2::BIGINT, $3::VARCHAR(64), $4::VARCHAR(40), $5::VARCHAR(32), $6::INTEGER, $7::INTEGER)",
			unicode.ToValidUTF8(fileStat.Name()), // name
			fileStat.Size(),                      // size
			hex.EncodeToString(f.Sha256[:]),      // sha256
			hex.EncodeToString(f.Sha1[:]),        // sha1
			hex.EncodeToString(f.Md5[:]),         // md5
			f.SymLinkInt(),                       // symlink
			f.NamedPipeInt(),                     // fifo
		).Scan(&faid); err != nil {
			err = errors.Wrapf(err, "error inserting file + file_alias")
			return err
		}

		// Insert file_belongs_archive
		if _, err = db.Exec("INSERT INTO file_belongs_archive(archive_id, file_id, path) "+
			"VALUES ($1, $2, $3) "+
			"ON CONFLICT (archive_id, file_id, path) DO NOTHING",
			arch.ArchiveID, // archive_id
			faid,           // file_id
			unicode.ToValidUTF8(strings.TrimPrefix(filePath, parentDirectory)), // path
		); err != nil {
			err = errors.Wrapf(err, "error inserting file_belongs_archive")
			return err
		}

		if err := StoreFile(blobStorage, f, filePath); err != nil {
			err = errors.Wrapf(err, "error storing %s", filePath)
			return err
		}

		return nil
	}
	// processFunc traverses the parent directory.
	// If it finds a directory, it is added to the priority queue.
	// If it finds an archive, it is extracted and recursively processed.
	// Otherwise it has found a file, and loads it into the database.
	processFunc := func(packageGraph *analysis.PackageGraph, parent string, fi os.FileInfo, parentNode *analysis.PackageNode) error {
		pth := filepath.Join(parent, fi.Name())
		if fi.IsDir() {
			// Add to priority queue
			d, err := NewHeapDir(pth)
			if err != nil {
				err = errors.Wrapf(err, "error making heap dir")
				return err
			}
			dq.Push(d)
		} else if rec := extract.IsExtractable(pth); rec == 1.0 {
			// If path looks extractable,
			// Process archive and add to arhive_contains_archive
			sub, err := InitArchive(pth, fi.Name())
			if err != nil {
				return err
			}
			if err := p.SyncArchive(db, sub); err != nil {
				return err
			}
			if !sub.FileCollectionID.Valid {
				// extract archive
				extDir := "/opt/tk/uploads/ext" // TODO make this non-static
				archiveExtractDir := extDir
				for i := 0; i < len(sub.Sha256.String)-2; i += 2 { // make a directory every two characters (1 byte) of the sha256
					archiveExtractDir = filepath.Join(archiveExtractDir, sub.Sha256.String[i:i+2])
				}
				if err := os.MkdirAll(archiveExtractDir, 0755); err != nil {
					err = errors.Wrapf(err, "error making archiveExtractDir %s", archiveExtractDir)
					return err
				}
				e, err := extract.NewAt(sub.Path.String, fi.Name(), archiveExtractDir)
				if err != nil {
					err = errors.Wrapf(err, "error setting-up extraction")
					return err
				}
				extractPath, err := e.Extract()
				if err != nil {
					// Treat as file instead
					return processAsFileFunc(db, sub.Path.String)
				}

				node, err := packageGraph.InsertHexString(fi.Name(), sub.Path.String, int(fi.Size()), sub.Sha1.String, sub.Sha256.String)
				if err != nil {
					return err
				}
				collectionNode.SubPackages.Add(node)
				if node.IsInCycle() {
					// skip package in cycle
					return nil
				}

				// process files
				if _, _, err := p.processFileCollection(db, packageGraph, node, sub, extractPath, blobStorage); err != nil {
					return err
				}
			}

			if _, err = db.Exec("INSERT INTO archive_contains(parent_id, child_id, path) "+
				"VALUES ($1, $2, $3) "+
				"ON CONFLICT(parent_id, child_id, path) DO NOTHING",
				arch.ArchiveID,       // parent_id
				sub.FileCollectionID, // child_id
				pth,                  // path
			); err != nil {
				err = errors.Wrapf(err, "error inserting into archive_contains(%d, %d, %s)", arch.ArchiveID, sub.FileCollectionID.Int64, pth)
				return err
			}
		} else {
			if err := processAsFileFunc(db, pth); err != nil {
				return err
			}
		}

		return nil
	}

	// Process initial files
	for _, v := range rootInfo {
		fi, err := v.Info()
		if err != nil {
			return nil, nil, err
		}

		if err = processFunc(packageGraph, parentDirectory, fi, collectionNode); err != nil {
			return nil, nil, err
		}
	}

	// Work through directory queue
	for len(dq) > 0 {
		d := dq.Pop().(*HeapDir)
		dlist, err := os.ReadDir(d.Path)
		if err != nil {
			err = errors.Wrapf(err, "error reading directory")
			return nil, nil, err
		}
		for _, v := range dlist {
			fi, err := v.Info()
			if err != nil {
				return nil, nil, err
			}

			if fi.Mode().IsRegular() || fi.IsDir() { // only process regular files or directories
				if err = processFunc(packageGraph, d.Path, fi, collectionNode); err != nil {
					return nil, nil, err
				}
			}
		}
	}

	// fetch verification code
	vcodeOne, vcodeTwo, err = p.CalculateArchiveVerificationCode(arch.ArchiveID)
	if err != nil {
		return nil, nil, err
	}

	// if err = db.QueryRowx("SELECT calculate_archive_verification_code($1)", arch.ID).Scan(&vcode); err != nil {
	// 	err = errors.Wrapf(err, "error getting verification_code from database")
	// 	return "", err
	// }

	// Upsert file_collection
	var cid int64
	if err = db.QueryRowx("INSERT INTO file_collection (verification_code_one, verification_code_two) VALUES ($1, $2) "+
		"ON CONFLICT (verification_code_one) DO UPDATE SET verification_code_two=EXCLUDED.verification_code_two "+
		"RETURNING id", vcodeOne, vcodeTwo).Scan(&cid); err != nil {
		return vcodeOne, vcodeTwo, errors.Wrapf(err, "error upserting file collection (\"%x\", \"%x\")", vcodeOne, vcodeTwo)
	}

	// Update archives
	db.NamedExec("UPDATE archive SET file_collection_id=:cid WHERE id=:aid", map[string]interface{}{
		"cid": cid,
		"aid": arch.ArchiveID,
	})

	// Transfer file_belongs_archive rows to file_belongs_collection
	if _, err = db.Exec("SELECT assign_archive_to_file_collection($1, $2)",
		arch.ArchiveID, // archive_id
		cid,            // file_collection_id
	); err != nil {
		err = errors.Wrapf(err, "error assigning archive file to file_collectior")
		return vcodeOne, vcodeTwo, err
	}

	arch.FileCollectionID.Int64 = cid
	arch.FileCollectionID.Valid = true

	return vcodeOne, vcodeTwo, nil
}
