-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
-- +goose StatementEnd

---- file
-- add new columns
ALTER TABLE file ADD license_expression TEXT;
ALTER TABLE file ADD license_notice TEXT;
ALTER TABLE file ADD copyright TEXT;

-- transfer old columns
-- (there will probably be none)
ALTER TABLE file ADD license_rationale TEXT;

-- +goose StatementBegin
DO LANGUAGE plpgsql $$
    DECLARE
        _row RECORD;
        _license_expression TEXT;
    BEGIN
        FOR _row IN SELECT file_id, license_id, rationale FROM file_license
        LOOP
            SELECT expression FROM license_expression WHERE id=_row.license_id INTO _license_expression;
            UPDATE file SET license_expression=_license_expression, license_rationale=_row.rationale WHERE id=_row.file_id;
        END LOOP;
    END;
$$;
-- +goose StatementEnd

-- drop
DROP TABLE IF EXISTS file_license;

---- file_collection
-- add new columns
ALTER TABLE file_collection ADD license_expression TEXT;
ALTER TABLE file_collection ADD license_notice TEXT;
ALTER TABLE file_collection ADD copyright TEXT;

-- transfer old columns
-- +goose StatementBegin
DO LANGUAGE plpgsql $$
    DECLARE
        _row RECORD;
        _license_expression TEXT;
    BEGIN
        FOR _row IN SELECT id, license_id FROM file_collection WHERE license_id IS NOT NULL
        LOOP
            SELECT expression FROM license_expression WHERE id=_row.license_id INTO _license_expression;
            IF _license_expression <> '' THEN
                UPDATE file_collection SET license_expression=_license_expression WHERE id=_row.id;
            END IF;
        END LOOP;
    END;
$$;
-- +goose StatementEnd

-- edit trigger function to keep new columns up-to-date with old columns
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION license_expression_update() RETURNS TRIGGER LANGUAGE plpgsql AS $$
    DECLARE
        _expression TEXT;
    BEGIN
        IF NEW.expression IS NULL
            OR OLD.license_id<>NEW.license_id
            OR OLD.left_id<>NEW.left_id
            OR OLD.right_id<>NEW.right_id
        THEN
            EXECUTE 'SELECT build_license_expression($1)'
                INTO _expression
                using NEW.id;
            UPDATE license_expression SET expression=_expression WHERE id=NEW.id;
            UPDATE license_expression SET expression=NULL WHERE left_id=NEW.id OR right_id=NEW.id;

            -- Update new license columns if applicable
            UPDATE file_collection SET license_expression=_expression WHERE license_id=NEW.id;
        END IF;

        RETURN NEW;
    END;
$$;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd

---- file
ALTER TABLE file DROP COLUMN license_expression;
ALTER TABLE file DROP COLUMN license_notice;
ALTER TABLE file DROP COLUMN copyright;

ALTER TABLE file DROP COLUMN license_rationale;

-- recreate empty file_license table
-- should've been empty in the first place anyways
CREATE TABLE IF NOT EXISTS file_license (
    id BIGSERIAL PRIMARY KEY,
    file_id BIGINT NOT NULL REFERENCES file(id),
    license_id BIGINT NOT NULL REFERENCES license_expression(id),
    rationale TEXT,
    insert_date TIMESTAMP NOT NULL DEFAULT NOW()
);

---- file_collection
ALTER TABLE file_collection DROP COLUMN license_expression;
ALTER TABLE file_collection DROP COLUMN license_notice;
ALTER TABLE file_collection DROP COLUMN copyright;

-- restore trigger function
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION license_expression_update() RETURNS TRIGGER LANGUAGE plpgsql AS $$
    DECLARE
        _expression TEXT;
    BEGIN
        IF NEW.expression IS NULL
            OR OLD.license_id<>NEW.license_id
            OR OLD.left_id<>NEW.left_id
            OR OLD.right_id<>NEW.right_id
        THEN
            EXECUTE 'SELECT build_license_expression($1)'
                INTO _expression
                using NEW.id;
            UPDATE license_expression SET expression=_expression WHERE id=NEW.id;
            UPDATE license_expression SET expression=NULL WHERE left_id=NEW.id OR right_id=NEW.id;
        END IF;

        RETURN NEW;
    END;
$$;
-- +goose StatementEnd