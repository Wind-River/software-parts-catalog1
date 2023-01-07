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
    FOR _file IN SELECT * FROM file_table WHERE _file.checksum_sha256 IS NOT NULL ORDER BY id
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
    FOR _file IN SELECT * FROM file_alias_table WHERE _file.checksum_sha256 IS NOT NULL ORDER BY id
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

-- copy archives and enter into new archive_alias table
-- +goose StatementBegin
DO LANGUAGE plpgsql $$
DECLARE
    _archive archive_table$ROWTYPE;
BEGIN
    FOR _archive IN SELECT * FROM archive_table WHERE _archive.checksum_sha256 IS NOT NULL ORDER BY id
    LOOP
        INSERT INTO archive (sha256, archive_size, md5, sha1, insert_date, storage_path) VALUES (_archive.checksum_sha256, _archive.archive_size, _archive.checksum_md5, _archive.checksum_sha1, _archive.insert_date, _archive.storage_path);
        INSERT INTO archive_alias (archive_sha256, name) VALUES (_archive.checksum_sha256, _archive.name);
    END LOOP;
END;
$$;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd
