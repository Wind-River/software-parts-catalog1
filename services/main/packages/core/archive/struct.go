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
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
	"wrs/tk/packages/blob"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"gitlab.devstar.cloud/ip-systems/extract.git"
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

		if err := p.process(m.archive, m.archive.Name.String); err != nil {
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

	if arch.ArchiveID == -1 {
		if err := p.SyncArchive(p.DB, arch); err != nil {
			return err
		}
	}

	ret := make(chan error)
	log.Trace().Str(zerolog.CallerFieldName, "ArchiveController.Process").Msg("Queueing message")
	p.queue <- archiveMessage{
		arch,
		ret,
	}

	log.Trace().Str(zerolog.CallerFieldName, "ArchiveController.Process").Msg("Waiting for return")
	defer os.Remove(arch.Path.String)
	return <-ret
}

// process is the function the goroutine usse to process the archive.
// The given archive is extracted, and the resulting files are loaded into the database.
func (p *ArchiveController) process(arch *Archive, fileName string) error {
	////
	// extract archive
	extDir := "/opt/tk/uploads/ext" // TODO make this non-static
	archiveExtractDir := filepath.Join(extDir, arch.Sha256.String)
	if err := os.MkdirAll(archiveExtractDir, 0755); err != nil {
		err = errors.Wrapf(err, "error making %s", archiveExtractDir)
		return err
	}
	extractor, err := extract.NewAt(arch.Path.String, // arch.Path.String
		fileName, // Filename
		archiveExtractDir,
	)
	if err != nil {
		return err
	}

	extractPath, err := extractor.Extract()
	if err != nil {
		return err
	}
	////
	defer func(path string) {
		log.Debug().Str("path", path).Msg("deferred cleaning up extracted directory")
		if err := os.RemoveAll(path); err != nil {
			log.Error().Err(err).Str("path", path).Msg("error cleaning up extracted directory")
		}
	}(archiveExtractDir)

	// process files
	vcodeOne, vcodeTwo, err := p.ProcessFileCollection(arch, extractPath, p.fileStorage)
	if err != nil {
		return err
	}
	log.Info().Bytes("vcodeOne", vcodeOne).Bytes("vcodeTwo", vcodeTwo).Str("sha256", arch.Sha256.String).Msg("ProcessedFileCollection")

	return nil
}

// SyncArchive upserts the given local archive into the database.
// If the archive is not already in S3, it is also stored at this time.
func (p *ArchiveController) SyncArchive(db *sqlx.DB, localArchive *Archive) error {
	log.Debug().Interface("local archive", localArchive).Str(zerolog.CallerFieldName, "archive/processor/s3.go:ArchiveProcessor.SyncArchive()").Send()
	var remoteArchive Archive

	// Upsert archive inherent values
	if err := db.QueryRowx("INSERT INTO archive(name, size, checksum_md5, checksum_sha1, checksum_sha256, extract_status) "+
		"VALUES($1, $2, $3, $4, $5, 0) "+
		"ON CONFLICT (name, checksum_sha1) "+
		"DO UPDATE SET checksum_md5=EXCLUDED.checksum_md5, checksum_sha256=EXCLUDED.checksum_sha256 "+
		"RETURNING *",
		localArchive.Name,
		localArchive.Size.Int64,
		localArchive.Md5.String,
		localArchive.Sha1,
		localArchive.Sha256.String,
	).StructScan(&remoteArchive); err != nil {
		err = errors.Wrapf(err, "error scanning remote archive")
		return err
	}

	log.Trace().Interface("remoteArchive", remoteArchive).Interface("localArchive", localArchive).Msg("UPSERTED archive")

	// If not valid, or does not exist, move source and update remote
	if !remoteArchive.Path.Valid || remoteArchive.Path.String == "" { // remote path does not exist or is invalid
		if err := p.S3Copy(localArchive, &remoteArchive); err != nil {
			return err
		}

		// Update archive path entry
		if _, err := db.Exec("UPDATE archive SET path=$1 WHERE id=$2", remoteArchive.Path.String, remoteArchive.ArchiveID); err != nil {
			err = errors.Wrapf(err, "error updating archive path")
			return err
		}
	}

	// Copy relevant fields from remote to local
	localArchive.FileCollectionID = remoteArchive.FileCollectionID
	localArchive.ArchiveID = remoteArchive.ArchiveID
	localArchive.InsertDate = remoteArchive.InsertDate
	localArchive.ExtractStatus = remoteArchive.ExtractStatus

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

	f, err := os.Open(localArchive.Path.String)
	if err != nil {
		return errors.Wrapf(err, "error opening local archive")
	}
	defer f.Close()

	key := filepath.Join("archive", localArchive.Sha256.String)

	if _, err := client.PutObject(&s3.PutObjectInput{
		Bucket: aws.String(p.bucket),
		Key:    aws.String(key),
		Body:   f,
	}); err != nil {
		return errors.Wrap(err, "error putting local archive")
	}

	remoteArchive.Path.String = fmt.Sprintf("s3://%s/%s", p.bucket, localArchive.Sha256.String)
	remoteArchive.Path.Valid = true

	return nil
}

type Archive struct {
	ArchiveID        int64          `db:"id"`
	FileCollectionID sql.NullInt64  `db:"file_collection_id"`
	Name             sql.NullString `db:"name"`
	Path             sql.NullString `db:"path"`
	Size             sql.NullInt64  `db:"size"`
	Sha1             sql.NullString `db:"checksum_sha1"`
	Sha256           sql.NullString `db:"checksum_sha256"`
	Md5              sql.NullString `db:"checksum_md5"`
	InsertDate       time.Time      `db:"insert_date"`
	ExtractStatus    int            `db:"extract_status"`
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
	ret.ArchiveID = -1
	ret.Name.String = name
	ret.Name.Valid = true
	ret.Path.String = source // Leave !Valid as it is not final location
	ret.Size.Int64 = size
	ret.Size.Valid = true
	ret.Md5.String = fmt.Sprintf("%x", md5)
	ret.Md5.Valid = true
	ret.Sha1.String = fmt.Sprintf("%x", sha1)
	ret.Sha1.Valid = true
	ret.Sha256.String = fmt.Sprintf("%x", sha256)
	ret.Sha256.Valid = true

	return ret, nil
}

// archiveMessage is used internally to send an archive to a goroutine and receive the error, or nil, result.
type archiveMessage struct {
	archive       *Archive
	returnChannel chan error
}
