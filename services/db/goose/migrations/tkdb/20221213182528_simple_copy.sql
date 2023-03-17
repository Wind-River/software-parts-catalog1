-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
-- +goose StatementEnd

-- copy files
-- +goose StatementBegin
DO LANGUAGE plpgsql $$
DECLARE
    _file file_table%ROWTYPE;
BEGIN
    FOR _file IN SELECT * FROM file_table WHERE checksum_sha256 IS NOT NULL ORDER BY id
    LOOP
        INSERT INTO file (sha256, file_size, md5, sha1) 
        VALUES (
                decode(_file.checksum_sha256, 'hex'), 
                _file.size, 
                decode(_file.checksum_md5, 'hex'), 
                decode(_file.checksum_sha1, 'hex')
        );
    END LOOP;
END;
$$;
-- +goose StatementEnd

-- copy file_aliases
-- +goose StatementBegin
DO LANGUAGE plpgsql $$
DECLARE
    _record RECORD;
    _sha1 BYTEA;
BEGIN
    FOR _record IN 
    SELECT file_alias_table.name, decode(file_table.checksum_sha256, 'hex') as sha256 
    FROM file_alias_table 
    INNER JOIN file_table ON file_table.id=file_alias_table.file_id 
    WHERE file_table.checksum_sha256 IS NOT NULL ORDER BY file_alias_table.id
    LOOP        
        INSERT INTO file_alias (file_sha256, name) VALUES (_record.sha256, _record.name);
    END LOOP;
END;
$$;
-- +goose StatementEnd

-- copy archives and enter into new archive_alias table
-- +goose StatementBegin
DO LANGUAGE plpgsql $$
DECLARE
    _archive archive_table%ROWTYPE;
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
