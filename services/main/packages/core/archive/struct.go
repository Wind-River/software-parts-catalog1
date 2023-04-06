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
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
	"wrs/tk/packages/array/hash"
	"wrs/tk/packages/blob"
	"wrs/tk/packages/blob/file"
	"wrs/tk/packages/core/archive/processor"
	"wrs/tk/packages/core/archive/sync"
	"wrs/tk/packages/core/archive/tree"
	"wrs/tk/packages/core/part"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/gabriel-vasile/mimetype"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type ArchiveController struct {
	DB *sqlx.DB

	fileStorage    blob.Storage
	archiveStorage blob.Storage

	maxThreads int
	queue      chan archiveMessage

	running bool

	awsConfig aws.Config
	session   *session.Session
	client    *s3.S3
	bucket    string
}

func NewArchiveController(db *sqlx.DB, fileStorage blob.Storage, archiveStorage blob.Storage, threads int, bucket string, credentials *credentials.Credentials, endpoint string, region string) *ArchiveController {
	ret := ArchiveController{
		DB:             db,
		fileStorage:    fileStorage,
		archiveStorage: archiveStorage,
		maxThreads:     threads,
		queue:          make(chan archiveMessage, threads),
		awsConfig: aws.Config{
			DisableSSL:       aws.Bool(true),
			S3ForcePathStyle: aws.Bool(true),
		},
		bucket: bucket,
	}

	if credentials != nil {
		ret.awsConfig.Credentials = credentials
	}
	if region != "" {
		ret.awsConfig.Region = aws.String(region)
	}
	if endpoint != "" {
		ret.awsConfig.Endpoint = aws.String(endpoint)
	}

	return &ret
}

// GetSession returns an AWS session, either new or cached.
func (p *ArchiveController) GetSession() (*session.Session, error) {
	if p.session == nil {
		sess, err := session.NewSession(&p.awsConfig)
		if err != nil {
			return nil, errors.Wrapf(err, "error creating session")
		}

		p.session = sess
	}

	return p.session, nil
}

// GetClient returns an S3 client, either new or cached.
func (p *ArchiveController) GetClient() (*s3.S3, error) {
	if p.client == nil {
		if _, err := p.GetSession(); err != nil {
			return nil, err
		}

		p.client = s3.New(p.session)
	}

	return p.client, nil
}

// Run checks that ArchiveController has what it needs to run, then starts the processing goroutines.
func (p *ArchiveController) Run() error {
	if p.DB == nil {
		err := errors.New("db is nil")
		return err
	}
	if p.fileStorage == nil {
		err := errors.New("blob storage is nil")
		return err
	}
	if p.maxThreads == 0 {
		p.maxThreads = 1
	}
	if p.queue == nil {
		p.queue = make(chan archiveMessage, p.maxThreads)
	}

	for i := 0; i < p.maxThreads; i++ {
		go p.run()
	}
	log.Info().
		Str(zerolog.CallerFieldName, "archive/processor.ArchiveController{}.Run()").
		Int("maxThreads", p.maxThreads).
		Msg("archive processor running")

	p.running = true

	return nil
}

// Close cleans-up the ArchiveController.
// The only thing currently closed is the queue channel.
func (p *ArchiveController) Close() error {
	close(p.queue)
	return nil
}

// run is the function used by the processing goroutines.
// It listens to the queue channel for an archive to process, then returns the result using the channel from the archiveMessage.
func (p *ArchiveController) run() error {
	defer func() {
		if err := recover(); err != nil {
			log.Info().Str(zerolog.CallerFieldName, "archive/processor.ArchiveController{}.run().recover()").Interface("error", err).Msg("gothread recovered and returning")
		}
	}()
	for {
		m, ok := <-p.queue
		if !ok {
			log.Info().Str(zerolog.CallerFieldName, "archive/processor.ArchiveController{}.run()").Err(nil).Msg("gothread returning")
			return nil // channel closed
		}
		log.Trace().Interface("message", m).Msg("Received message")

		if err := p.process(m.archive, m.archive.Aliases[0]); err != nil {
			m.returnChannel <- err
		}

		m.returnChannel <- nil
	}
}

// Process is the public API for ArchiveController that hides the underlying channel/actor interaction.
// The given archive will be sent to a goroutine with a return channel for extraction and processing.
func (p *ArchiveController) Process(arch *Archive) error {
	log.Trace().Interface("arch", arch).Msg("ArchiveController.Process")
	if !p.running {
		if err := p.Run(); err != nil {
			return err
		}
	}

	// if arch.ArchiveID == -1 { // TODO? determine if archive in database?
	if err := p.SyncArchive(p.DB, arch); err != nil {
		return err
	}
	// }

	ret := make(chan error)
	log.Trace().Str(zerolog.CallerFieldName, "ArchiveController.Process").Msg("Queueing message")
	p.queue <- archiveMessage{
		arch,
		ret,
	}

	log.Trace().Str(zerolog.CallerFieldName, "ArchiveController.Process").Msg("Waiting for return")
	defer os.Remove(arch.StoragePath.String) // should this value be re-used in this way for local storage?
	return <-ret
}

// TODOC
func (p *ArchiveController) visitArchive(archivePath string, archive *tree.Archive) error { // Visit Archive
	log.Debug().Str("archivePath", archivePath).Msg("Uploading Archive")

	var remoteArchive Archive
	// Upsert archive inherent values
	if err := p.DB.QueryRowx(`INSERT INTO archive (sha256, archive_size, md5, sha1)
	VALUES ($1, $2, $3, $4)
	ON CONFLICT (sha256) DO UPDATE SET sha256=EXCLUDED.sha256
	RETURNING *`,
		archive.Sha256[:], archive.Size, archive.Md5[:], archive.Sha1[:]).StructScan(&remoteArchive); err != nil {
		return errors.Wrapf(err, "error upserting archive")
	}

	// If not valid, or does not exist, move source and update remote
	if !remoteArchive.StoragePath.Valid || remoteArchive.StoragePath.String == "" { // remote path does not exist or is invalid
		if err := p.S3Copy(&Archive{
			Sha256: hash.Sha256(archive.Sha256),
			StoragePath: sql.NullString{
				Valid:  false,
				String: archivePath,
			},
		}, &remoteArchive); err != nil {
			return err
		}

		// Update archive path entry
		if _, err := p.DB.Exec("UPDATE archive SET storage_path=$1 WHERE sha256=$2", remoteArchive.StoragePath.String, remoteArchive.Sha256); err != nil {
			err = errors.Wrapf(err, "error updating archive path")
			return err
		}
	}

	return nil
}

// TODOC
func (p *ArchiveController) visitFile(filePath string, f *tree.File) error { // Visit file
	log.Debug().Str("filePath", filePath).Msg("Uploading File")
	// Only store if file is a normal file greater than 0 bytes
	if f.Size < 1 {
		return nil
	}

	mimeType, err := mimetype.DetectFile(filePath)
	if err != nil {
		err = errors.Wrap(err, "error detecting mimetype")
		return err
	}

	r, err := os.Open(filePath)
	if err != nil {
		err = errors.Wrapf(err, "error opening file")
		return err
	}
	defer r.Close()

	if err := p.fileStorage.Store(r, &file.FileInfo{
		Sha256:   file.Sha256(f.Sha256),
		Sha1:     file.Sha1(f.Sha1),
		Size:     f.Size,
		MimeType: mimeType.String(),
	}); err != nil {
		return err
	}

	return nil
}

// process is the function the goroutine usse to process the archive.
// The given archive is extracted, and the resulting files are loaded into the database.
func (p *ArchiveController) process(arch *Archive, fileName string) error {
	log.Debug().Interface("arch", arch).Str(zerolog.CallerFieldName, "ArchiveController.process").Msg("About to process archive")
	ap, err := processor.NewArchiveProcessor(
		p.visitArchive,
		p.visitFile,
	)
	if err != nil {
		return err
	}
	rootArchive, err := ap.ProcessArchive(arch.StoragePath.String, nil)
	if err != nil {
		return err
	}
	rootArchive.Name = fileName
	if err := tree.CalculateVerificationCodes(rootArchive); err != nil {
		return err
	}
	log.Debug().Interface("arch", arch).Str(zerolog.CallerFieldName, "ArchiveController.process").Msg("Created archive tree")

	partID, err := sync.SyncTree(p.DB, &part.PartController{DB: p.DB}, rootArchive) // TODO properly obtain part controller
	if err != nil {
		return err
	}
	log.Debug().Interface("arch", arch).Str(zerolog.CallerFieldName, "ArchiveController.process").Str("partID", partID.String()).Msg("Synced archive tree")

	return nil
}

// SyncArchive upserts the given local archive into the database.
// If the archive is not already in S3, it is also stored at this time.
func (p *ArchiveController) SyncArchive(db *sqlx.DB, localArchive *Archive) error {
	log.Debug().Interface("local archive", localArchive).Str(zerolog.CallerFieldName, "archive/processor/s3.go:ArchiveProcessor.SyncArchive()").Send()
	var remoteArchive Archive

	// Upsert archive inherent values
	if err := db.QueryRowx(`INSERT INTO archive(sha256, archive_size, md5, sha1) 
	VALUES ($1, $2, $3, $4) 
	ON CONFLICT (sha256) DO UPDATE SET sha256=EXCLUDED.sha256
	RETURNING *`, // Meaningless update is required, if no insert or update is made RETURNING will return nothing
		localArchive.Sha256, localArchive.Size, localArchive.Md5, localArchive.Sha1).StructScan(&remoteArchive); err != nil {
		return errors.Wrapf(err, "error scanning remote archive")
	}

	log.Trace().Interface("remoteArchive", remoteArchive).Interface("localArchive", localArchive).Msg("UPSERTED archive")

	// If not valid, or does not exist, move source and update remote
	if !remoteArchive.StoragePath.Valid || remoteArchive.StoragePath.String == "" { // remote path does not exist or is invalid
		if err := p.S3Copy(localArchive, &remoteArchive); err != nil {
			return err
		}

		// Update archive path entry
		if _, err := db.Exec("UPDATE archive SET storage_path=$1 WHERE sha256=$2", remoteArchive.StoragePath.String, remoteArchive.Sha256); err != nil {
			err = errors.Wrapf(err, "error updating archive path")
			return err
		}
	}

	// Copy relevant fields from remote to local
	localArchive.PartID = remoteArchive.PartID
	// localArchive.ArchiveID = remoteArchive.ArchiveID // set exists in db
	localArchive.InsertDate = remoteArchive.InsertDate
	// localArchive.ExtractStatus = remoteArchive.ExtractStatus

	log.Trace().Interface("archive", localArchive).Msg("Synced Archive")

	return nil
}

// S3Copy copies a local archive to S3
func (p *ArchiveController) S3Copy(localArchive *Archive, remoteArchive *Archive) error {
	log.Debug().Interface("local archive", localArchive).Interface("remote archive", remoteArchive).Str(zerolog.CallerFieldName, "archive/processor/s3.go:ArchiveProcessor.s3Move()").Send()
	client, err := p.GetClient()
	if err != nil {
		return err
	}

	f, err := os.Open(localArchive.StoragePath.String)
	if err != nil {
		return errors.Wrapf(err, "error opening local archive")
	}
	defer f.Close()

	key := filepath.Join("archive", hex.EncodeToString(localArchive.Sha256[:]))

	if _, err := client.PutObject(&s3.PutObjectInput{
		Bucket: aws.String(p.bucket),
		Key:    aws.String(key),
		Body:   f,
	}); err != nil {
		return errors.Wrap(err, "error putting local archive")
	}

	remoteArchive.StoragePath.String = fmt.Sprintf("s3://%s/%x", p.bucket, localArchive.Sha256[:])
	remoteArchive.StoragePath.Valid = true

	return nil
}

type Archive struct {
	Sha256      hash.Sha256    `db:"sha256"`
	Size        int64          `db:"archive_size"`
	PartID      *part.ID       `db:"part_id"`
	Md5         hash.Md5       `db:"md5"`
	Sha1        hash.Sha1      `db:"sha1"`
	InsertDate  time.Time      `db:"insert_date"`
	StoragePath sql.NullString `db:"storage_path"`
	Aliases     []string       `db:"names"`
}

// InitArchive loads an Archive from the local file system.
// All data necessary for inserting into the database is calculated at this time.
func InitArchive(source string, name string) (*Archive, error) {
	// Verify that source exists
	stat, err := os.Stat(source)
	if err != nil {
		err = errors.Wrapf(err, "could not stat source")
		return nil, err
	}

	// Get name from filepath if necessary
	if name == "" {
		name = filepath.Base(source)
	}

	// Get size
	size := stat.Size()

	// Calculate hashes
	md5Hasher := md5.New()
	sha1Hasher := sha1.New()
	sha256Hasher := sha256.New()

	f, err := os.Open(source)
	if err != nil {
		err = errors.Wrapf(err, "could not open file")
		return nil, err
	}
	defer f.Close()

	for {
		buf := make([]byte, 64) // block size
		n, err := f.Read(buf)

		if err == io.EOF {
			break
		} else if err != nil {
			err = errors.Wrapf(err, "error chunking file")
			return nil, err
		}

		buf = buf[:n] // set slice to size actually read

		if _, err := md5Hasher.Write(buf); err != nil {
			err = errors.Wrapf(err, "error calculating md5")
			return nil, err
		}
		if _, err := sha1Hasher.Write(buf); err != nil {
			err = errors.Wrapf(err, "error calculating sha1")
			return nil, err
		}
		if _, err := sha256Hasher.Write(buf); err != nil {
			err = errors.Wrapf(err, "error calculating sha256")
			return nil, err
		}
	}

	md5 := md5Hasher.Sum(nil)
	sha1 := sha1Hasher.Sum(nil)
	sha256 := sha256Hasher.Sum(nil)

	ret := new(Archive)
	// ret.ArchiveID = -1
	ret.Aliases = []string{name}
	ret.StoragePath.String = source // Leave !Valid as it is not final location
	ret.Size = size
	copy(ret.Md5[:], md5)
	copy(ret.Sha1[:], sha1)
	copy(ret.Sha256[:], sha256)

	return ret, nil
}

// archiveMessage is used internally to send an archive to a goroutine and receive the error, or nil, result.
type archiveMessage struct {
	archive       *Archive
	returnChannel chan error
}
