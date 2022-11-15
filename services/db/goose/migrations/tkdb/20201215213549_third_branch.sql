-- +goose Up
--
--

CREATE TABLE IF NOT EXISTS archive (
    id BIGSERIAL PRIMARY KEY,
    file_collection_id BIGINT REFERENCES file_collection(id),
    name TEXT,
    path TEXT,
    size INTEGER,
    checksum_sha1 VARCHAR(40),
    UNIQUE(name, checksum_sha1),
    checksum_sha256 VARCHAR(64),
    checksum_md5 VARCHAR(32),
    insert_date TIMESTAMP NOT NULL DEFAULT NOW(),
    extract_status INTEGER NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS archive_contains (
    parent_id BIGINT NOT NULL REFERENCES archive(id),
    child_id BIGINT NOT NULL REFERENCES file_collection(id),
    path TEXT NOT NULL,
    PRIMARY KEY(parent_id, child_id, path)
);

CREATE TABLE IF NOT EXISTS file_belongs_archive (
    archive_id BIGINT NOT NULL REFERENCES archive(id),
    file_id BIGINT NOT NULL REFERENCES file_alias(id),
    path TEXT NOT NULL,
    PRIMARY KEY(archive_id, file_id, path)
);

CREATE TABLE IF NOT EXISTS crypto_record_action (
    action_id BIGSERIAL PRIMARY KEY,
    crypto_record_id BIGINT NOT NULL REFERENCES file_have_crypto_record(id),
    action TEXT NOT NULL,
    user_identifier TEXT NOT NULL DEFAULT ''::TEXT,
    date TIMESTAMP NOT NULL DEFAULT NOW(),
    value TEXT
);

CREATE TABLE IF NOT EXISTS crypto_records_review (
    crypto_record_id BIGINT NOT NULL REFERENCES file_have_crypto_record(id),
    useraccount_id BIGINT NOT NULL DEFAULT 0,
    comments TEXT,
    response TEXT
);

--
--

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd
