-- +goose Up
--
--

CREATE TABLE IF NOT EXISTS analyst (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    job_title TEXT,
    note TEXT,
    insert_date TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS composite_component (
    id BIGSERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    description TEXT
);

CREATE TABLE IF NOT EXISTS component_contains_component (
    parent_component BIGINT NOT NULL REFERENCES composite_component(id),
    child_component BIGINT NOT NULL REFERENCES composite_component(id),
    CHECK(parent_component <> child_component)
);

CREATE TABLE IF NOT EXISTS file (
    id BIGSERIAL PRIMARY KEY,
    checksum_sha1 VARCHAR(40) NOT NULL UNIQUE,
    checksum_sha256 VARCHAR(64) UNIQUE,
    checksum_md5 VARCHAR(32) UNIQUE,
    insert_date TIMESTAMP DEFAULT NOW(),
    flag_symlink INTEGER NOT NULL DEFAULT 0,
    flag_fifo INTEGER NOT NULL DEFAULT 0,
    size BIGINT NOT NULL DEFAULT -1::INTEGER
);

CREATE TABLE IF NOT EXISTS file_alias (
    id BIGSERIAL PRIMARY KEY,
    file_id BIGINT REFERENCES file(id),
    name VARCHAR(400),
    UNIQUE(file_id, name)
);

CREATE TABLE IF NOT EXISTS group_container (
    id BIGSERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    type VARCHAR(20),
    associatedlicense TEXT,
    associatedrationale TEXT,
    description TEXT,
    comments TEXT,
    insert_date TIMESTAMP NOT NULL DEFAULT NOW(),
    parent_id BIGINT,
    UNIQUE (name, parent_id)
);

CREATE TABLE IF NOT EXISTS groups (
    group_id BIGSERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    type VARCHAR(20),
    associatedlicense TEXT,
    associatedrationale TEXT,
    date_created TIMESTAMP,
    description TEXT,
    comments TEXT
);

-- +goose StatementBegin
DO LANGUAGE plpgsql $$
BEGIN
    IF NOT EXISTS (SELECT * FROM pg_type typ INNER JOIN pg_namespace nsp ON nsp.oid = typ.typnamespace WHERE nsp.nspname = current_schema() AND typ.typname = 'op_type')
    THEN
        CREATE TYPE op_type AS ENUM (
            'INSERT',
            'UPDATE',
            'DELETE'
        );
    END IF;
END;
$$;
-- +goose StatementEnd

CREATE TABLE IF NOT EXISTS history (
    target TEXT NOT NULL,
    useraccount_id BIGINT NOT NULL,
    operation op_type,
    unique_key JSONB,
    value JSONB,
    time TIMESTAMP NOT NULL DEFAULT NOW(),
    PRIMARY KEY (target, useraccount_id, time)
);

CREATE TABLE IF NOT EXISTS license (
    id BIGSERIAL PRIMARY KEY,
    name TEXT DEFAULT '',
    type VARCHAR(32) NOT NULL DEFAULT 'custom',
    identifier TEXT NOT NULL UNIQUE,
    UNIQUE(name, identifier),
    text TEXT,
    insert_date TIMESTAMP NOT NULL DEFAULT NOW()
);

-- +goose StatementBegin
CREATE OR REPLACE FUNCTION license_cascade_update() RETURNS TRIGGER LANGUAGE plpgsql AS $$
    BEGIN
        UPDATE license_expression SET expression=NULL WHERE license_id=NEW.id;
        RETURN NEW;
    END;
$$;
-- +goose StatementEnd

DROP TRIGGER IF EXISTS license_trigger ON license;
CREATE TRIGGER license_trigger AFTER UPDATE ON license FOR EACH ROW EXECUTE PROCEDURE license_cascade_update();

--
--

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd
