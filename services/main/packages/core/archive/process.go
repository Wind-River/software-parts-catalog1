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

	collectionNode, err := packageGraph.InsertHexString(archive.Aliases[0], archive.StoragePath.String, int(archive.Size), hex.EncodeToString(archive.Sha1[:]), hex.EncodeToString(archive.Sha256[:]))
	if err != nil {
		return nil, nil, err
	}

	return p.processFileCollection(p.DB, packageGraph, collectionNode,
		archive, parentDirectory, blobStorage)
}

// TODO WSTRPG-86; assigning files
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

		if !fileStat.Mode().IsRegular() {
			return nil // skip irregular file
		}

		f, err := NewFile(filePath)
		if err != nil {
			return err
		}

		// Insert file
		if _, err = db.Exec(`INSERT INTO file (sha256, file_size, md5, sha1, label) 
		VALUES ($1, $2, $3, $4, $5) ON CONFLICT (sha256) DO NOTHING`,
			f.Sha256[:],
			fileStat.Size(),
			f.Md5[:],
			f.Sha1[:],
			fileStat.Name()); err != nil {
			return errors.Wrapf(err, "error inserting file")
		}

		// Insert file_alias
		if _, err = db.Exec("INSERT INTO file_alias (file_sha256, name) VALUES ($1, $2) ON CONFLICT (file_sha256, name) DO NOTHING",
			f.Sha256[:], fileStat.Name()); err != nil {
			return errors.Wrapf(err, "error inserting file_alias")
		}

		// Insert file_belongs_archive
		if _, err = db.Exec("INSERT INTO file_belongs_archive(archive_id, file_id, path) "+
			"VALUES ($1, $2, $3) "+
			"ON CONFLICT (archive_id, file_id, path) DO NOTHING",
			"TODO", // archive_id
			"TODO", // file_id
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
			if sub.PartID == nil {
				// extract archive
				extDir := "/opt/tk/uploads/ext" // TODO make this non-static
				archiveExtractDir := extDir
				for i := 0; i < len(sub.Sha256); i++ { // make a directory every two characters (1 byte) of the sha256
					character := hex.EncodeToString(sub.Sha256[i : i+1])
					archiveExtractDir = filepath.Join(archiveExtractDir, character)
				}
				if err := os.MkdirAll(archiveExtractDir, 0755); err != nil {
					err = errors.Wrapf(err, "error making archiveExtractDir %s", archiveExtractDir)
					return err
				}
				e, err := extract.NewAt(sub.StoragePath.String, fi.Name(), archiveExtractDir)
				if err != nil {
					err = errors.Wrapf(err, "error setting-up extraction")
					return err
				}
				extractPath, err := e.Extract()
				if err != nil {
					// Treat as file instead
					return processAsFileFunc(db, sub.StoragePath.String)
				}

				node, err := packageGraph.InsertHexString(fi.Name(), sub.StoragePath.String, int(fi.Size()), hex.EncodeToString(sub.Sha1[:]), hex.EncodeToString(sub.Sha256[:]))
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
				arch.Sha256, // parent_id
				sub.PartID,  // child_id
				pth,         // path
			); err != nil {
				err = errors.Wrapf(err, "error inserting into archive_contains(%x, %s, %s)", arch.Sha256, *sub.PartID, pth)
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
	vcodeOne, vcodeTwo, err = p.CalculateArchiveVerificationCode(404) // arch.ArchiveID
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
		"aid": arch.Sha256,
	})

	// Transfer file_belongs_archive rows to file_belongs_collection
	if _, err = db.Exec("SELECT assign_archive_to_file_collection($1, $2)",
		arch.Sha256, // archive_id
		cid,         // file_collection_id
	); err != nil {
		err = errors.Wrapf(err, "error assigning archive file to file_collectior")
		return vcodeOne, vcodeTwo, err
	}

	// arch.PartID = cid set partid

	return vcodeOne, vcodeTwo, nil
}
