-- +goose Up

CREATE TABLE IF NOT EXISTS partlist(
   id BIGSERIAL PRIMARY KEY,
   name TEXT NOT NULL,
   parent_id BIGINT REFERENCES partlist(id) ON DELETE CASCADE
);

--
--

CREATE TABLE IF NOT EXISTS partlist_has_part(
   partlist_id BIGINT NOT NULL REFERENCES partlist(id) ON DELETE CASCADE,
   part_id UUID NOT NULL REFERENCES part(part_id),
   PRIMARY KEY(partlist_id, part_id)
);

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd
