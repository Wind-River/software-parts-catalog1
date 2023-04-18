-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
-- +goose StatementEnd
ALTER TABLE part ADD COLUMN IF NOT EXISTS label TEXT;
ALTER TABLE part DROP COLUMN IF EXISTS license_notice;
ALTER TABLE part DROP COLUMN IF EXISTS automation_license;
ALTER TABLE part DROP COLUMN IF EXISTS automation_license_rationale;

ALTER TABLE file ADD COLUMN IF NOT EXISTS label TEXT;
-- +goose StatementBegin
DO LANGUAGE plpgsql $$
    DECLARE
        _sha256 SHA256_BYTEA;
        _alias TEXT;
    BEGIN
        FOR _sha256 IN SELECT sha256 FROM file WHERE label IS NULL
        LOOP
            SELECT name FROM file_alias WHERE file_sha256=_sha256 LIMIT 1 INTO _alias;
            UPDATE file SET label=_alias WHERE sha256=_sha256;
        END LOOP;
    END
$$;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd
ALTER TABLE part DROP COLUMN IF EXISTS label;
ALTER TABLE part ADD COLUMN IF NOT EXISTS license_notice TEXT;
ALTER TABLE part ADD COLUMN IF NOT EXISTS automation_license TEXT;
ALTER TABLE part ADD COLUMN IF NOT EXISTS automation_license_rationale TEXT;

ALTER TABLE file DROP COLUMN IF EXISTS label;