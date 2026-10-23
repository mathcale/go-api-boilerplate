DROP INDEX IF EXISTS users_email_idx;

CREATE UNIQUE INDEX IF NOT EXISTS users_email_idx ON users (email);
