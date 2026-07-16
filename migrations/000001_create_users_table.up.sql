-- pgcrypto provides gen_random_uuid(); it is enabled here so the users table can
-- default its primary key without the application supplying an id.
CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE IF NOT EXISTS users (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email               TEXT UNIQUE NOT NULL,
    password            TEXT NOT NULL,
    name                TEXT,
    surname             TEXT,
    phone               TEXT,
    roles               TEXT[] NOT NULL DEFAULT '{}',
    verification_code   TEXT,
    verification_expiry TIMESTAMPTZ,
    activated           BOOLEAN NOT NULL DEFAULT FALSE
);
