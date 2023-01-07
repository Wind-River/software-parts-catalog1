-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
-- +goose StatementEnd

-- rename old tables needed for later temporary migration steps
ALTER TABLE file RENAME TO file_table;
ALTER TABLE file_alias RENAME TO file_alias_table;
ALTER TABLE archive RENAME TO archive_table;

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd