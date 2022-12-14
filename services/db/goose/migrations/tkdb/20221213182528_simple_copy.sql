-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
-- +goose StatementEnd

-- +goose StatementBegin
-- copy files
DO LANGUAGE plpgsql $$
DECLARE
    _file file$ROWTYPE;
BEGIN
    FOR _file IN SELECT * FROM file
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

-- copy file_aliases

-- copy archives

-- copy archive_aliases


-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd
