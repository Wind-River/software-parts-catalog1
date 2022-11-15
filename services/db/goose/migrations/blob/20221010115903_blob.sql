-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
-- +goose StatementEnd

CREATE TABLE IF NOT EXISTS blob_metadata (
    sha256 BYTEA PRIMARY KEY,
    sha1 BYTEA NOT NULL,
    size BIGINT NOT NULL,
    mime TEXT
);

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd

DROP TABLE IF EXISTS blob_metadata;
