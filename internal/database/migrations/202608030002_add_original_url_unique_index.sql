-- +goose Up
DELETE FROM shortened_urls first
USING shortened_urls second
WHERE first.original_url = second.original_url
    AND first.id > second.id;

CREATE UNIQUE INDEX IF NOT EXISTS shortened_urls_original_url_idx
    ON shortened_urls (original_url);

-- +goose Down
DROP INDEX IF EXISTS shortened_urls_original_url_idx;
