-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
-- +goose StatementEnd

-- rename old tables needed for later temporary migration steps
ALTER TABLE file RENAME TO file_table;
ALTER TABLE file_alias RENAME TO file_alias_table;
ALTER TABLE archive RENAME TO archive_table;
ALTER TABLE archive_alias RENAME TO archive_alias_table;

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd

ALTER TABLE file_table RENAME TO file;
ALTER TABLE file_alias_table RENAME TO file_alias;
ALTER TABLE archive_table RENAME TO archive;
ALTER TABLE archive_alias_table RENAME TO archive_alias;
