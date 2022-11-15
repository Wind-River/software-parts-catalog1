-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
-- +goose StatementEnd

ALTER TABLE file_collection ADD COLUMN verification_code_one BYTEA UNIQUE;
-- set FVC1\x00 || verification_code
UPDATE file_collection SET verification_code_one=(decode('4656433100', 'hex') || decode(verification_code, 'hex')) where verification_code is not null;
ALTER TABLE file_collection DROP COLUMN verification_code;

-- +goose StatementBegin
CREATE OR REPLACE FUNCTION stub_file_collection(vcode BYTEA) RETURNS BIGINT LANGUAGE plpgsql AS $$
    DECLARE
        ret BIGINT;
    BEGIN
        SELECT id FROM file_collection WHERE verification_code_one=vcode INTO ret;
        IF ret IS NULL THEN
            INSERT INTO file_collection (verification_code_one) VALUES (vcode) RETURNING id INTO ret;
        END IF;

        RETURN ret;
    END;
$$;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd
