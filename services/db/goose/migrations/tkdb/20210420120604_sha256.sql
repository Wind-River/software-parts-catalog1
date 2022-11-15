-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
-- +goose StatementEnd

-- +goose StatementBegin
CREATE OR REPLACE FUNCTION insert_file(_name TEXT, _sha256 VARCHAR(64), _sha1 VARCHAR(40), _md5 VARCHAR(32), _symlink INTEGER) RETURNS BIGINT LANGUAGE plpgsql AS $$
    DECLARE
        _fid BIGINT;
    BEGIN
        INSERT INTO file (checksum_sha256, checksum_sha1, checksum_md5, flag_symlink)
        VALUES (_sha256, _sha1, _md5, _symlink)
        ON CONFLICT (checksum_sha1) DO UPDATE SET checksum_md5=EXCLUDED.checksum_md5, checksum_sha256=EXCLUDED.checksum_sha256
        RETURNING id
        INTO _fid;

        IF _fid IS NOT DISTINCT FROM NULL THEN
            RAISE EXCEPTION 'Insert into raw file table did not return an ID';
            RETURN NULL;
        END IF;

        INSERT INTO file_alias (file_id, name) VALUES (_fid, _name)
        ON CONFLICT (file_id, name) DO UPDATE SET name=EXCLUDED.name
        RETURNING id
        INTO _fid;

        RETURN _fid;
    END;
$$;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE OR REPLACE FUNCTION insert_file(_name TEXT, _sha256 VARCHAR(64), _sha1 VARCHAR(40), _md5 VARCHAR(32), _symlink INTEGER, _fifo INTEGER) RETURNS BIGINT LANGUAGE plpgsql AS $$
    DECLARE
        _fid BIGINT;
    BEGIN
        INSERT INTO file (checksum_sha256, checksum_sha1, checksum_md5, flag_symlink, flag_fifo)
        VALUES (_sha256, _sha1, _md5, _symlink, _fifo)
        ON CONFLICT (checksum_sha1) DO UPDATE SET checksum_md5=EXCLUDED.checksum_md5, checksum_sha256=EXCLUDED.checksum_sha256
        RETURNING id
        INTO _fid;

        IF _fid IS NOT DISTINCT FROM NULL THEN
            RAISE EXCEPTION 'Insert into raw file table did not return an ID';
            RETURN NULL;
        END IF;

        INSERT INTO file_alias (file_id, name) VALUES (_fid, _name)
        ON CONFLICT (file_id, name) DO UPDATE SET name=EXCLUDED.name
        RETURNING id
        INTO _fid;

        RETURN _fid;
    END;
$$;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE OR REPLACE FUNCTION insert_file(_name TEXT, _size BIGINT, _sha256 VARCHAR(64), _sha1 VARCHAR(40), _md5 VARCHAR(32), _symlink INTEGER, _fifo INTEGER) RETURNS BIGINT LANGUAGE plpgsql AS $$
    DECLARE
        _fid BIGINT;
    BEGIN
        INSERT INTO file (size, checksum_sha256, checksum_sha1, checksum_md5, flag_symlink, flag_fifo)
        VALUES (_size, _sha256, _sha1, _md5, _symlink, _fifo)
        ON CONFLICT (checksum_sha1) DO UPDATE SET checksum_md5=EXCLUDED.checksum_md5, size=EXCLUDED.size, checksum_sha256=EXCLUDED.checksum_sha256
        RETURNING id
        INTO _fid;

        IF _fid IS NOT DISTINCT FROM NULL THEN
            RAISE EXCEPTION 'Insert into raw file table did not return an ID';
            RETURN NULL;
        END IF;

        INSERT INTO file_alias (file_id, name) VALUES (_fid, _name)
        ON CONFLICT (file_id, name) DO UPDATE SET name=EXCLUDED.name
        RETURNING id
        INTO _fid;

        RETURN _fid;
    END;
$$;
-- +goose StatementEnd


-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd
