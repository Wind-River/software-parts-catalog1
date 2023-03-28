-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
-- +goose StatementEnd

-- make csv of data about to be deleted
COPY (
    SELECT a.name as archive_name, a.size as archive_size, a.checksum_md5 as md5, a.checksum_sha1 as sha1, a.checksum_sha256 as sha256,
    fc.verification_code_one, fc.verification_code_two,
    (SELECT build_group_path(fc.group_container_id)) as family_name,
    le.expression as license_expression, fc.license_rationale,
    a.insert_date as archive_insert_date, fc.insert_date as file_collection_insert_date, a.path as storage_path, a.extract_status as archive_extract_status
    FROM archive a
    LEFT JOIN file_collection fc ON fc.id=a.file_collection_id
    LEFT JOIN license_expression le ON le.id=fc.license_id
) TO '/tmp/archive_data_as-is.csv' WITH CSV DELIMITER ',' HEADER;

-- delete collections of files that don't have sha256s
-- TODO

-- rename old tables needed for later temporary migration steps
ALTER TABLE file RENAME TO file_table;
ALTER TABLE file_alias RENAME TO file_alias_table;
ALTER TABLE archive RENAME TO archive_table;

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd