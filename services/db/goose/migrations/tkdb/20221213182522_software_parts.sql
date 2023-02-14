-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
-- +goose StatementEnd
-- Extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "ltree";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- Domain Types for Hashes
CREATE DOMAIN sha256_bytea AS BYTEA CHECK (OCTET_LENGTH(VALUE) = 32);
CREATE DOMAIN sha1_bytea AS BYTEA CHECK (OCTET_LENGTH(VALUE) = 20);
CREATE DOMAIN md5_bytea AS BYTEA CHECK (OCTET_LENGTH(VALUE) = 16);

-- Part
CREATE TABLE IF NOT EXISTS part (
    part_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    type LTREE,
    name TEXT,
    version TEXT,
    family_name TEXT,
    file_verification_code BYTEA UNIQUE,
    size BIGINT,
    license TEXT,
    license_rationale TEXT,
    automation_license TEXT,
    comprised UUID REFERENCES part(part_id)
);

-- +goose StatementBegin
CREATE OR REPLACE FUNCTION calculate_part_verification_code_v2(_pid UUID) RETURNS BYTEA LANGUAGE plpgsql AS $$
    DECLARE
        _row RECORD;
        _shas SHA256_BYTEA[] := ARRAY[]::SHA256_BYTEA[];
        _data BYTEA := '';
        _vcode BYTEA;
    BEGIN
        FOR _row IN SELECT f.sha256 FROM file f
            INNER JOIN part_has_file phf ON phf.file_sha256=f.sha256
            INNER JOIN part p ON p.part_id=phf.part_id
            WHERE p.part_id = _pid
            AND f.sha256 IS NOT NULL
        LOOP
            _shas := _shas || _row.sha256;
        END LOOP;

        FOR _row IN WITH RECURSIVE parts AS (
                SELECT php.child_id as part_id FROM part_has_part php WHERE php.parent_id=_pid
                UNION
                SELECT child_id as part_id FROM part_has_part
                INNER JOIN parts ON parts.part_id=part_has_part.parent_id
            )
            SELECT f.sha256 FROM file f
            INNER JOIN part_has_file phf ON phf.file_sha256=f.sha256
            INNER JOIN part p ON p.part_id=phf.part_id
            INNER JOIN parts ON parts.part_id=p.part_id
            WHERE f.sha256 IS NOT NULL
        LOOP
            _shas := _shas || _row.sha256;
        END LOOP;

        FOR _row IN SELECT UNNEST(_shas) AS s ORDER BY s
        LOOP
            _data := _data || _row.s;
        END LOOP;

        SELECT digest(_data, 'sha256') INTO _vcode;

        RETURN '\x4656433200'::BYTEA || _vcode;
    END;
$$;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE OR REPLACE FUNCTION verify_part_modification() RETURNS TRIGGER LANGUAGE plpgsql AS $$
    DECLARE
        _vcode BYTEA;
    BEGIN
        IF NEW.file_verification_code IS NULL THEN
            RETURN NEW; -- no verification code to verify
        END IF;

        IF OLD.file_verification_code = NEW.file_verification_code THEN
            RETURN NEW; -- verification code was not changed
        END IF;

        SELECT calculate_part_verification_code_v2(NEW.part_id) INTO _vcode;

        IF _vcode <> NEW.file_verification_code THEN
            RAISE EXCEPTION 'Declared file_verification_code "%" does not match Calculated file_verification_code "%"', NEW.file_verification_code, _vcode;
        END IF;

        RETURN NEW;
    END;
$$;
-- +goose StatementEnd

CREATE TRIGGER verify_part_file_verification_code_trigger BEFORE INSERT OR UPDATE ON part FOR EACH ROW EXECUTE PROCEDURE verify_part_modification();

CREATE TABLE IF NOT EXISTS part_has_part (
    parent_id UUID REFERENCES part(part_id),
    child_id UUID REFERENCES part(part_id),
    path TEXT,
    PRIMARY KEY(parent_id, child_id, path),
);

CREATE TABLE IF NOT EXISTS part_alias (
    alias TEXT PRIMARY KEY,
    part_id UUID REFERENCES part(part_id)
);

-- Part Documents
CREATE TABLE IF NOT EXISTS part_has_document (
    part_id UUID REFERENCES part(part_id),
    key TEXT NOT NULL,
    PRIMARY KEY(part_id, key),
    document JSONB NOT NULL
);

CREATE TABLE IF NOT EXISTS part_documents (
    part_id UUID REFERENCES part(part_id),
    key TEXT NOT NULL,
    title TEXT NOT NULL,
    PRIMARY KEY(part_id, key, title),
    document JSONB NOT NULL
);

-- Files
CREATE TABLE IF NOT EXISTS file (
    sha256 SHA256_BYTEA PRIMARY KEY,
    file_size BIGINT NOT NULL DEFAULT 0,
    md5 MD5_BYTEA,
    sha1 SHA1_BYTEA
);

CREATE TABLE IF NOT EXISTS file_alias (
    file_sha256 SHA256_BYTEA REFERENCES file(sha256),
    name TEXT NOT NULL,
    PRIMARY KEY(file_sha256, name)
);

CREATE TABLE IF NOT EXISTS part_has_file (
    part_id UUID REFERENCES part(part_id),
    file_sha256 SHA256_BYTEA REFERENCES file(sha256),
    PATH TEXT NOT NULL,
    PRIMARY KEY(part_id, file_sha256, path)
);

-- File Documents
CREATE TABLE IF NOT EXISTS file_has_document (
    file_sha256 SHA256_BYTEA REFERENCES file(sha256),
    key TEXT NOT NULL,
    PRIMARY KEY(file_sha256, key),
    document JSONB NOT NULL
);

CREATE TABLE IF NOT EXISTS file_documents (
    file_sha256 SHA256_BYTEA REFERENCES file(sha256),
    key TEXT NOT NULL,
    title TEXT NOT NULL,
    PRIMARY KEY(file_sha256, key, title),
    document JSONB NOT NULL
);

-- Archive
CREATE TABLE IF NOT EXISTS archive (
    sha256 BYTEA PRIMARY KEY,
    archive_size BIGINT NOT NULL DEFAULT 0,
    part_id UUID REFERENCES part(part_id),
    md5 MD5_BYTEA,
    sha1 SHA1_BYTEA,
    insert_date TIMESTAMP NOT NULL DEFAULT NOW(),
    storage_path TEXT
);

CREATE TABLE IF NOT EXISTS archive_alias (
    archive_sha256 SHA256_BYTEA REFERENCES archive(sha256),
    name TEXT NOT NULL,
    PRIMARY KEY(archive_sha256, name)
);

CREATE TABLE IF NOT EXISTS archive_has_archive (
    parent_sha256 BYTEA REFERENCES archive(sha256),
    child_sha256 BYTEA REFERENCES archive(sha256),
    CHECK(parent_sha256<>child_sha256),
    path TEXT
);
-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd
