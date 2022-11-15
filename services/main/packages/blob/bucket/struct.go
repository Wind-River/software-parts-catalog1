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
	"bytes"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

// Initialization of the BlobBucket struct

type BlobBucket struct {
	db     *sqlx.DB
	prefix string
	// Bucket
	awsConfig aws.Config
	session   *session.Session
	client    *s3.S3
	bucket    string
}

// NewBlobBucket initializes a BlobBucket struct
func NewBlobBucket(db *sqlx.DB, prefix string, bucket string, region string, endpoint string, credentials *credentials.Credentials) *BlobBucket {
	log.Info().Str("db", fmt.Sprintf("%p", db)).Str("prefix", prefix).Str("bucket", bucket).Str("region", region).Str("endpoint", endpoint).Str("credentials", fmt.Sprintf("%p", credentials)).Msg("NewBlobBucket")
	ret := new(BlobBucket)
	ret.db = db
	ret.prefix = prefix
	ret.bucket = bucket

	ret.awsConfig = aws.Config{
		DisableSSL:       aws.Bool(true),
		S3ForcePathStyle: aws.Bool(true),
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

	// test s3
	client, err := ret.GetClient()
	if err != nil {
		log.Warn().Err(err).Str("region", region).Str("endpoint", endpoint).Str("credentials", fmt.Sprintf("%p", credentials)).Msg("Failed to Get Client")
	}

	output, err := client.ListBuckets(&s3.ListBucketsInput{})
	if err != nil {
		log.Warn().Err(err).Str("region", region).Str("endpoint", endpoint).Str("credentials", fmt.Sprintf("%p", credentials)).Msg("Failed to List Buckets")
	}
	for i, b := range output.Buckets {
		log.Debug().Int("index", i).Str("bucket", b.GoString()).Msg("Listing Buckets")
	}

	if _, err := client.PutObject(&s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String("letters.txt"),
		Body:   bytes.NewReader([]byte{'a', 'b', 'c'}),
	}); err != nil {
		log.Warn().Err(err).Str("bucket", bucket).Str("region", region).Str("endpoint", endpoint).Str("credentials", fmt.Sprintf("%p", credentials)).Msg("Failed to PutObject")
	}

	return ret
}

// GetSession creates, if necessary, and returns an aws.Session
func (bucket *BlobBucket) GetSession() (*session.Session, error) {
	if bucket.session == nil {
		sess, err := session.NewSession(&bucket.awsConfig)
		if err != nil {
			return nil, errors.Wrapf(err, "error creating session")
		}

		bucket.session = sess
	}

	return bucket.session, nil
}

// GetClient creates, if necessary, and returns an s3 client
func (bucket *BlobBucket) GetClient() (*s3.S3, error) {
	if bucket.client == nil {
		if _, err := bucket.GetSession(); err != nil {
			return nil, err
		}

		bucket.client = s3.New(bucket.session)
	}

	return bucket.client, nil
}

// CreateBlobBucket creates the metadata table if necessary, and initializes the blob bucket struct
func CreateBlobBucket(db *sqlx.DB, prefix string, bucket string, region string, endpoint string, credentials *credentials.Credentials) (*BlobBucket, error) {
	log.Trace().Str("bucket", bucket).Str("region", region).Str("endpoint", endpoint).Msg("Creating Blob Bucket")
	// if _, err := db.Exec(`
	// 	CREATE TABLE IF NOT EXISTS blob_metadata (
	// 		sha256 BYTEA PRIMARY KEY,
	// 		sha1 BYTEA NOT NULL,
	// 		size BIGINT NOT NULL,
	// 		mime TEXT
	// 	)
	// `); err != nil {
	// 	err = errors.Wrapf(err, "error creating blob table")
	// 	return nil, err
	// }

	blobBucket := NewBlobBucket(db, prefix, bucket, region, endpoint, credentials)

	return blobBucket, nil
}
