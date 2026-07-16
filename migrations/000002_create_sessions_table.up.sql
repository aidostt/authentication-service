CREATE TABLE IF NOT EXISTS sessions (
    userid        UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    refresh_token TEXT NOT NULL,
    expires_at    TIMESTAMPTZ NOT NULL
);

-- One session row per user: SetSession upserts on this key. The unique index is
-- also what ON CONFLICT (userid) targets.
CREATE UNIQUE INDEX IF NOT EXISTS sessions_userid_key ON sessions (userid);

-- GetByRefreshToken looks sessions up by token on every refresh.
CREATE INDEX IF NOT EXISTS sessions_refresh_token_idx ON sessions (refresh_token);
