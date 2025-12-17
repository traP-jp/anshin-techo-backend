-- +goose up

ALTER TABLE users
  DROP COLUMN id,
  DROP COLUMN created_at,
  DROP COLUMN updated_at,
  ADD PRIMARY KEY (traq_id);