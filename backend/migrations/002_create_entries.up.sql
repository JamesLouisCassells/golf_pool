CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE entries (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    year INT NOT NULL,
    clerk_id TEXT REFERENCES users(clerk_id),
    display_name TEXT NOT NULL,
    picks JSONB NOT NULL,
    in_overs BOOLEAN NOT NULL DEFAULT false,
    locked BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
