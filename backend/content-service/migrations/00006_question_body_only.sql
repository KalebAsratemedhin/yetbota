-- +goose Up
-- Collapse question title into body; questions use a single body field.
UPDATE questions
SET body = CASE
    WHEN coalesce(trim(body), '') = '' THEN title
    WHEN coalesce(trim(title), '') = '' THEN body
    ELSE title || E'\n\n' || body
END;

ALTER TABLE questions DROP COLUMN IF EXISTS title;

-- +goose Down
ALTER TABLE questions ADD COLUMN IF NOT EXISTS title TEXT NOT NULL DEFAULT '';
UPDATE questions SET title = left(body, 300) WHERE coalesce(title, '') = '';
