-- +goose Up
ALTER TABLE users
    ADD COLUMN google_id TEXT,
    ADD COLUMN display_name TEXT,
    ADD COLUMN avatar_url TEXT,
    ALTER COLUMN password_hash DROP NOT NULL,
    ALTER COLUMN password_salt DROP NOT NULL;

CREATE INDEX IF NOT EXISTS idx_users_google_id ON users(google_id);

-- +goose Down
DROP INDEX IF EXISTS idx_users_google_id;

ALTER TABLE users
    ALTER COLUMN password_hash SET NOT NULL,
    ALTER COLUMN password_salt SET NOT NULL,
    DROP COLUMN IF EXISTS google_id,
    DROP COLUMN IF EXISTS display_name,
    DROP COLUMN IF EXISTS avatar_url;
