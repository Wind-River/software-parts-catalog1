-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
-- +goose StatementEnd

-- copy files
-- +goose StatementBegin
DO LANGUAGE plpgsql $$
DECLARE
    _file file_table$ROWTYPE;
BEGIN
    FOR _file IN SELECT * FROM file_table ORDER BY id
    LOOP
        INSERT INTO file (sha256, file_size, md5, sha1) VALUES (_file.checksum_sha256, _file.size, _file.checksum_md5, _file.checksum_sha1);
    END LOOP;
END;
$$;
-- +goose StatementEnd

-- copy file_aliases
-- +goose StatementBegin
DO LANGUAGE plpgsql $$
DECLARE
    _file_alias file_alias_table$ROWTYPE;
    _sha1 BYTEA;
BEGIN
    FOR _file IN SELECT * FROM file_alias_table ORDER BY id
    LOOP
        SELECT _file.checksum_sha1 
        FROM file_table 
        WHERE file_table.id=_file_alias.id 
        INTO _sha1;
        
        INSERT INTO file_alias (file_sha256, name) VALUES (_file.checksum_sha256, _file.size, _file.checksum_md5, _file.checksum_sha1);
    END LOOP;
END;
$$;
-- +goose StatementEnd

-- copy archives

-- copy archive_aliases


-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd
