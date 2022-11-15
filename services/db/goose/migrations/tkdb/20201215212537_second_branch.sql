-- +goose Up
--
--

CREATE TABLE IF NOT EXISTS file_collection (
    id BIGSERIAL PRIMARY KEY,
    verification_code VARCHAR(40) UNIQUE,
    insert_date TIMESTAMP NOT NULL DEFAULT NOW(),
    group_container_id INTEGER REFERENCES group_container(id),
    flag_extract INTEGER NOT NULL DEFAULT 0,
    flag_license_extracted INTEGER NOT NULL DEFAULT 0,
    license_id BIGINT REFERENCES license_expression(id),
    license_rationale TEXT,
    analyst_id BIGINT REFERENCES analyst(id)
);

CREATE TABLE IF NOT EXISTS file_collection_contains (
    parent_id BIGINT NOT NULL REFERENCES file_collection(id),
    child_id BIGINT NOT NULL REFERENCES file_collection(id) CHECK (parent_id<>child_id),
    path TEXT NOT NULL,
    PRIMARY KEY(parent_id, child_id, path)
);

CREATE TABLE IF NOT EXISTS file_belongs_collection (
    file_collection_id BIGINT NOT NULL REFERENCES file_collection(id),
    file_id BIGINT NOT NULL REFERENCES file(id),
    path TEXT NOT NULL,
    PRIMARY KEY(file_collection_id, file_id, path)
);

CREATE TABLE IF NOT EXISTS component_contains_collection (
    component_id BIGINT NOT NULL REFERENCES composite_component(id),
    file_collection_id BIGINT NOT NULL REFERENCES file_collection(id),
    PRIMARY KEY(component_id, file_collection_id)
);

-- +goose StatementBegin
DO LANGUAGE plpgsql $$
BEGIN
    IF NOT EXISTS (SELECT * FROM pg_type typ INNER JOIN pg_namespace nsp ON nsp.oid = typ.typnamespace WHERE nsp.nspname = current_schema() AND typ.typname = 'crypto_prediction')
    THEN
        CREATE TYPE crypto_prediction AS (
            prediction BOOLEAN,
            probability DOUBLE PRECISION
        );
    END IF;
END;
$$;
-- +goose StatementEnd

-- +goose StatementBegin
DO LANGUAGE plpgsql $$
BEGIN
    IF NOT EXISTS (SELECT * FROM pg_type typ INNER JOIN pg_namespace nsp ON nsp.oid = typ.typnamespace WHERE nsp.nspname = current_schema() AND typ.typname = 'review_type')
    THEN
        CREATE TYPE review_type AS ENUM (
            'false-positive',
            'TBD',
            'positive'
        );
    END IF;
END;
$$;
-- +goose StatementEnd

CREATE TABLE IF NOT EXISTS file_have_crypto_evidence (
    file_id BIGINT NOT NULL REFERENCES file(id),
    type TEXT NOT NULL,
    PRIMARY KEY(file_id, type),
    useraccount_id BIGINT NOT NULL DEFAULT 0,
    review REVIEW_TYPE NOT NULL,
    comment TEXT,
    prediction CRYPTO_PREDICTION,
    crypto_record_log TEXT,
    crypto_record_hash VARCHAR(40) UNIQUE
);

CREATE TABLE IF NOT EXISTS file_have_crypto_record (
    id BIGSERIAL PRIMARY KEY,
    method TEXT NOT NULL,
    file_id BIGINT NOT NULL REFERENCES file(id),
    math_type TEXT NOT NULL,
    match_text TEXT NOT NULL,
    match_line_number INTEGER NOT NULL,
    match_file_index_begin INTEGER NOT NULL,
    match_file_index_end INTEGER NOT NULL,
    match_line_index_begin INTEGER NOT NULL,
    match_line_index_end INTEGER NOT NULL,
    line_text TEXT NOT NULL,
    line_text_before_1 TEXT,
    line_text_before_2 TEXT,
    line_text_before_3 TEXT,
    line_text_after_1 TEXT,
    line_text_after_2 TEXT,
    line_text_after_3 TEXT,
    human_reviewed TEXT,
    comments TEXT,
    crypto_record_log TEXT
);

--
--

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd
