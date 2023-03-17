-- remove rest of unneeded tkdb tables

-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
-- +goose StatementEnd

DROP TABLE IF EXISTS analyst CASCADE;
DROP TABLE IF EXISTS archive_contains;
DROP TABLE IF EXISTS archive_has_archive;
DROP TABLE IF EXISTS archive_table CASCADE;
DROP TABLE IF EXISTS component_contains_collection;
DROP TABLE IF EXISTS component_contains_component;
DROP TABLE IF EXISTS composite_component;
DROP TABLE IF EXISTS crypto_record_action;
DROP TABLE IF EXISTS crypto_records_review;
DROP TABLE IF EXISTS file_alias_table CASCADE;
DROP TABLE IF EXISTS file_belongs_archive;
DROP TABLE IF EXISTS file_belongs_collection;
DROP TABLE IF EXISTS file_collection CASCADE;
DROP TABLE IF EXISTS file_collection_contains;
DROP TABLE IF EXISTS file_have_crypto_evidence;
DROP TABLE IF EXISTS file_have_crypto_record;
DROP TABLE IF EXISTS file_table;
DROP TABLE IF EXISTS group_container;
DROP TABLE IF EXISTS groups;
DROP TABLE IF EXISTS history;
DROP TABLE IF EXISTS license_expression;
DROP TABLE IF EXISTS license;

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd
