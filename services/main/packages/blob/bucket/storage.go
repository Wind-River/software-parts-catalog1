// Copyright (c) 2020 Wind River Systems, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at:
//       http://www.apache.org/licenses/LICENSE-2.0
// Unless required by applicable law or agreed to in writing, software  distributed
// under the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES
// OR CONDITIONS OF ANY KIND, either express or implied.

package bucket

import (
	"crypto/sha1"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"io"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
	"wrs/tk/packages/blob/file"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// Implementation of the methods required for blob.Storage

func (bucket BlobBucket) Store(data io.Reader, metadata *file.FileInfo) error {
	// log.Trace().Interface("metadata", metadata).Str("bucket", bucket.bucket).Msg("Storing File")
	if _, err := bucket.GetClient(); err != nil {
		return err
	}

	key := bucket.prefix
	// for _, v := range metadata.Sha256 {
	// 	key = filepath.Join(key, hex.EncodeToString([]byte{v}))
	// }
	key = filepath.Join(key, hex.EncodeToString(metadata.Sha256[:]))

	var info file.FileInfo
	var exists bool
	if err := bucket.db.QueryRowx("SELECT * FROM blob_metadata WHERE sha256=$1", info.Sha256).StructScan(&info); err != nil && err != sql.ErrNoRows {
		err = errors.Wrapf(err, "error checking if blob exists")
		return err
	} else if err == nil {
		exists = true
	}

	if exists {
		if info.Sha256.Hex() == metadata.Sha256.Hex() &&
			info.Sha1.Hex() == metadata.Sha1.Hex() &&
			info.Size == metadata.Size {
			// Verify remote file
			if bucket.hasObject(*metadata, key) {
				return nil
			}
		} // else overwrite existing
	}

	// determine which mimetype to use if any
	// due to past bugs, mimetype may or may not be stored
	var mimeType sql.NullString
	if info.MimeType != "" {
		mimeType.String = info.MimeType
		mimeType.Valid = true
	} else if metadata.MimeType != "" {
		mimeType.String = metadata.MimeType
		mimeType.Valid = true
	}

	uploader := s3manager.NewUploader(bucket.session)
	for tries := 0; ; tries++ {
		if _, err := uploader.Upload(&s3manager.UploadInput{
			Body:   data,
			Bucket: aws.String(bucket.bucket),
			Key:    aws.String(key),
		}); err != nil && tries >= 3 {
			return errors.Wrapf(err, "error uploading object to %s@%s", key, bucket.bucket)
		} else if err != nil {
			time.Sleep(time.Minute * time.Duration(math.Pow(2, float64(tries))))
			continue
		}

		break
	}

	// upsert
	if _, err := bucket.db.Exec("INSERT INTO blob_metadata(size, mime, sha256, sha1) VALUES ($1, $2, $3, $4) ON CONFLICT (sha256) DO UPDATE SET "+
		"sha1=EXCLUDED.sha1, size=EXCLUDED.size, mime=EXCLUDED.mime",
		metadata.Size, mimeType, metadata.Sha256, metadata.Sha1); err != nil {
		err = errors.Wrapf(err, "error inserting metadata")
		return err
	}

	return nil
}

func (bucket BlobBucket) hasObject(info file.FileInfo, key string) bool {
	client, err := bucket.GetClient()
	if err != nil {
		log.Error().Err(err).Msg("error getting client")
		return false
	}

	head, err := client.HeadObject(&s3.HeadObjectInput{
		Bucket: aws.String(bucket.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		log.Error().Err(err).Interface("info", info).Str("key", key).Msg("error s3 HeadObject")
		return false
	}

	if head == nil {
		log.Debug().Interface("info", info).Str("key", key).Msg("head is nil")
		return false
	}

	if _, ok := head.Metadata["Content-Length"]; !ok {
		log.Warn().Interface("info", info).Str("key", key).Interface("head", head).Interface("metadata", head.Metadata).Msg("No Content-Length")
		return true
	}

	remoteSize, err := strconv.ParseInt(*head.Metadata["Content-Length"], 10, 64)
	if err != nil {
		log.Error().Err(err).Interface("info", info).Str("key", key).Interface("metadata", head.Metadata).Msg("error parsing length")
		return false
	}

	if remoteSize != info.Size {
		log.Debug().Interface("info", info).Str("key", key).
			Int64("remote_size", remoteSize).Int64("expected_size", info.Size).
			Msg("size does not match")
		return false
	}

	// We can get Md5 from AWS, but we don't store that in the database to check
	// So we rely on path, which is based on sha256 anyways, and size
	return true
}

func (bucket BlobBucket) Retrieve(hash file.Sha256) (*file.File, error) {
	var info file.FileInfo
	if err := bucket.db.QueryRowx("SELECT * FROM blob_metadata WHERE sha256=$1", hash).StructScan(&info); err != nil {
		// err = errors.Wrapf(err, "error selecting %x", hash)
		// return nil, err
		info.Sha256 = hash
	}
	log.Trace().Interface("FileInfo", info).Msg("Scanned Struct")

	key := filepath.Join(bucket.prefix, hex.EncodeToString(info.Sha256[:])) // new key
	if bucket.hasObject(info, key) || true {
		return bucket.retrieve(info, key)
	} else {
		log.Debug().Interface("info", info).Str("key", key).Msg("Bucket does not have new key")
	}

	// use old key
	key = bucket.prefix
	for _, v := range info.Sha256 {
		key = filepath.Join(key, hex.EncodeToString([]byte{v}))
	}
	log.Debug().Str("key", key).Msg("Trying old key style")

	return bucket.retrieve(info, key)
}

func (bucket BlobBucket) retrieve(info file.FileInfo, key string) (*file.File, error) {
	sess, err := bucket.GetSession()
	if err != nil {
		return nil, err
	}

	// Create tmp file
	tmp, err := os.CreateTemp("", "BlobBucket-*")
	if err != nil {
		return nil, errors.Wrapf(err, "error creating tmp file")
	}
	// Download to tmp file
	downloader := s3manager.NewDownloader(sess)
	if _, err := downloader.Download(tmp, &s3.GetObjectInput{
		Bucket: aws.String(bucket.bucket),
		Key:    aws.String(key),
	}); err != nil {
		return nil, errors.Wrapf(err, "error downloading %s to %s", key, tmp.Name())
	}
	// seek tmp file
	if _, err := tmp.Seek(0, 0); err != nil {
		return nil, errors.Wrapf(err, "error seeking tmp file %s to start", tmp.Name())
	}

	return &file.File{
			FileInfo:       info,
			ReadSeekCloser: &TransientFile{tmp},
		},
		nil
}

func (bucket BlobBucket) ListAll() ([]file.FileInfo, error) {
	rows, err := bucket.db.Queryx("SELECT * FROM blob_metadata")
	if err != nil {
		err = errors.Wrapf(err, "error selecting all")
		return nil, err
	}
	defer rows.Close()

	ret := make([]file.FileInfo, 0)
	for rows.Next() {
		var tmp file.FileInfo
		if err := rows.StructScan(&tmp); err != nil {
			err = errors.Wrapf(err, "error scanning")
			return ret, err
		}

		ret = append(ret, tmp)
	}

	return ret, nil
}

func (bucket BlobBucket) StreamAll() (chan file.FileInfo, error) {
	rows, err := bucket.db.Queryx("SELECT * FROM blob_metadata")
	if err != nil {
		err = errors.Wrapf(err, "error selecting all")
		return nil, err
	}
	defer rows.Close()

	ch := make(chan file.FileInfo)

	go func(rows *sqlx.Rows, ch chan file.FileInfo) error {
		defer rows.Close()
		defer close(ch)

		for rows.Next() {
			var tmp file.FileInfo
			if err := rows.StructScan(&tmp); err != nil {
				err = errors.Wrapf(err, "error scanning")
				return err
			}

			ch <- tmp
		}

		return nil
	}(rows, ch)

	return ch, nil
}

// TODO remove
func (bucket BlobBucket) Migrate() error {
	logger := log.With().Str(zerolog.CallerFieldName, "BlobBucket.Migrate()").Logger()
	logger.Info().Send()

	client, err := bucket.GetClient()
	if err != nil {
		logger.Error().Err(err).Msg("error getting client")
		return err
	}

	// 1. list all objects
	// 2. migrate key if necessary
	// 3. upsert into database

	type Work struct {
		Sha256 [32]byte       `json:"sha256"`
		OldKey sql.NullString `json:"old_key"`
		NewKey sql.NullString `json:"new_key"`
	}

	var workList []Work

	if _, err := os.Stat("/opt/tk/blob/migration.json"); err == nil {
		f, err := os.Open("/opt/tk/blob/migration.json")
		if err != nil {
			logger.Error().Err(err).Msg("error loading past migration file")
		} else {
			if err := json.NewDecoder(f).Decode(&workList); err != nil {
				logger.Error().Err(err).Msg("error decoding past migration file")
			}
		}
		f.Close()
	}

	if len(workList) == 0 {
		workMap := make(map[[32]byte]*Work)
		if err := client.ListObjectsV2Pages(&s3.ListObjectsV2Input{
			Bucket: aws.String(bucket.bucket),
			Prefix: aws.String("blob/"),
		}, func(output *s3.ListObjectsV2Output, lastPage bool) (Continue bool) {
			for _, object := range output.Contents {
				key := strings.TrimPrefix(*object.Key, "blob/")
				logger.Trace().Str("key", key).Msg("Checking key")

				if key[2] == '/' { // ..(/) == old key
					sha256Slice, err := hex.DecodeString(strings.ReplaceAll(key, "/", ""))
					if err != nil {
						logger.Error().Err(err).Str("key", *object.Key).Msg("error decoding sha256")
						return false
					}
					var sha256 [32]byte
					copy(sha256[:], sha256Slice)

					if w, ok := workMap[sha256]; ok {
						w.OldKey.String = *object.Key
						w.OldKey.Valid = true
					} else {
						workMap[sha256] = &Work{
							Sha256: sha256,
							OldKey: sql.NullString{
								String: *object.Key,
								Valid:  true,
							},
						}
					}
				} else { // new key
					sha256Slice, err := hex.DecodeString(key)
					if err != nil {
						logger.Error().Err(err).Str("key", *object.Key).Msg("error decoding sha256")
						return false
					}
					var sha256 [32]byte
					copy(sha256[:], sha256Slice)

					if w, ok := workMap[sha256]; ok {
						w.NewKey.String = *object.Key
						w.NewKey.Valid = true
					} else {
						workMap[sha256] = &Work{
							Sha256: sha256,
							NewKey: sql.NullString{
								String: *object.Key,
								Valid:  true,
							},
						}
					}
				}
			}

			return !lastPage
		}); err != nil {
			logger.Error().Err(err).Msg("error listing objects for migration")
			return err
		}

		workList = make([]Work, 0, len(workMap))
		for _, w := range workMap {
			workList = append(workList, *w)
		}

		f, err := os.Create("/opt/tk/blob/migration.json")
		if err != nil {
			logger.Error().Err(err).Msg("error creating past migration file")
		} else {
			if err := json.NewEncoder(f).Encode(workList); err != nil {
				logger.Error().Err(err).Msg("error writing past migration file")
			}
		}
		f.Close()
	}

	logger.Info().Int("work count", len(workList)).Send()

	for _, w := range workList {
		// download blob
		f, err := bucket.Retrieve(w.Sha256)
		if err != nil {
			logger.Error().Err(err).Str("sha256", hex.EncodeToString(w.Sha256[:])).Msg("error retrieving file")
			continue // return err
		}
		defer f.Close()

		// calculate metadata
		f.FileInfo.Size = 0
		sha1Hasher := sha1.New()
		sha256Hasher := sha256.New()
		multiW := io.MultiWriter(sha1Hasher, sha256Hasher)
		if n, err := io.Copy(multiW, f); err != nil {
			logger.Error().Err(err).
				Str("sha256", hex.EncodeToString(w.Sha256[:])).
				Msg("error reading file")
			continue
		} else {
			f.FileInfo.Size = n
		}
		copy(f.FileInfo.Sha1[:], sha1Hasher.Sum(nil))
		copy(f.FileInfo.Sha256[:], sha256Hasher.Sum(nil))

		// upsert
		if _, err := bucket.db.Exec("INSERT INTO blob_metadata(size, sha256, sha1) VALUES ($1, $2, $3) ON CONFLICT (sha256) DO UPDATE SET "+
			"sha1=EXCLUDED.sha1, size=EXCLUDED.size",
			f.FileInfo.Size, f.FileInfo.Sha256, f.FileInfo.Sha1); err != nil {
			err = errors.Wrapf(err, "error inserting metadata")
			logger.Error().Err(err).Interface("f.FileInfo", f.FileInfo).Msg("upsert error")
			// return err
		}

		if !w.NewKey.Valid { // migrate
			f.Seek(0, 0)
			if err := bucket.Store(f, &f.FileInfo); err != nil {
				logger.Error().Err(err).Interface("metadata", f.FileInfo).Interface("work", w).Msg("error storing blob")
				continue // return err
			}
		}

		if w.OldKey.Valid { // delete
			if _, err := client.DeleteObject(&s3.DeleteObjectInput{
				Bucket: aws.String(bucket.bucket),
				Key:    aws.String(w.OldKey.String),
			}); err != nil {
				logger.Warn().Err(err).Interface("work", w).Msg("error deleting old object")
			}

			logger.Trace().Str("oldKey", w.OldKey.String).Msg("oldKey deleted")
		}
	}

	return nil
}
