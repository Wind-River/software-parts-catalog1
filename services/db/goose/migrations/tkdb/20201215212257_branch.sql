-- +goose Up
--
--

CREATE TABLE IF NOT EXISTS license_expression (
    id BIGSERIAL PRIMARY KEY,
    type TEXT NOT NULL DEFAULT 'LICENSE'::TEXT,
    license_id BIGINT REFERENCES license(id),
    expression TEXT,
    operator TEXT,
    left_id BIGINT REFERENCES license_expression(id),
    right_id BIGINT REFERENCES license_expression(id)
);

-- +goose StatementBegin
CREATE OR REPLACE FUNCTION build_license_expression(_id BIGINT) RETURNS TEXT LANGUAGE plpgsql AS $$
    DECLARE
        expression TEXT;
        rec RECORD;
        leftt TEXT := '';
        rightt TEXT := '';
    BEGIN
        SELECT le.type, l.identifier, le.operator, le.left_id, le.right_id INTO rec FROM license_expression le LEFT JOIN license l ON l.id=le.license_id WHERE le.id=_id;

        IF rec.type = 'LICENSE' THEN
            RETURN rec.identifier;
        ELSE
            expression = rec.operator;

            IF rec.right_id IS NOT NULL THEN
                SELECT build_license_expression(rec.right_id) INTO rightt;
            END IF;

            IF rec.left_id IS NOT NULL THEN
                SELECT build_license_expression(rec.left_id) INTO leftt;
            END IF;

            expression = leftt || ' ' || expression || ' ' || rightt;

            return expression;
        END IF;
    END;
$$;
-- +goose StatementEnd

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
DROP TRIGGER IF EXISTS license_expression_trigger ON license_expression;
CREATE TRIGGER license_expression_trigger AFTER INSERT OR UPDATE ON license_expression FOR EACH ROW EXECUTE PROCEDURE license_expression_update();

CREATE TABLE IF NOT EXISTS file_license (
    id BIGSERIAL PRIMARY KEY,
    file_id BIGINT NOT NULL REFERENCES file(id),
    license_id BIGINT NOT NULL REFERENCES license_expression(id),
    rationale TEXT,
    insert_date TIMESTAMP NOT NULL DEFAULT NOW()
);

--
--

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd
