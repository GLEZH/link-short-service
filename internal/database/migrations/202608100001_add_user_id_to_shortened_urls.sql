-- +goose Up
ALTER TABLE shortened_urls
    ADD COLUMN IF NOT EXISTS user_id TEXT NOT NULL DEFAULT '';

CREATE INDEX IF NOT EXISTS shortened_urls_user_id_idx
    ON shortened_urls (user_id);

-- +goose Down
DROP INDEX IF EXISTS shortened_urls_user_id_idx;

ALTER TABLE shortened_urls
    DROP COLUMN IF EXISTS user_id;
