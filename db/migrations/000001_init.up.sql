CREATE EXTENSION IF NOT EXISTS postgis;
CREATE EXTENSION IF NOT EXISTS citext;
CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE OR REPLACE FUNCTION set_updated_at()
RETURNS trigger AS $$
BEGIN
    NEW.updated_at = now();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TABLE users (
    id TEXT PRIMARY KEY CHECK (id ~ '^usr_[A-Za-z0-9]{8}$'),
    name VARCHAR(100) NOT NULL,
    email CITEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    profile_image_url TEXT,
    active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TRIGGER trg_users_updated_at
BEFORE UPDATE ON users
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();

CREATE TABLE programs (
    id TEXT PRIMARY KEY CHECK (id ~ '^prog_[A-Za-z0-9]{8}$'),
    merchant_id TEXT NOT NULL,
    merchant_name VARCHAR(200) NOT NULL,
    lat DOUBLE PRECISION NOT NULL CHECK (lat BETWEEN -90 AND 90),
    lng DOUBLE PRECISION NOT NULL CHECK (lng BETWEEN -180 AND 180),
    stamp_goal INTEGER NOT NULL CHECK (stamp_goal BETWEEN 1 AND 100),
    reward_name VARCHAR(200) NOT NULL,
    reward_image_url TEXT,
    reward_description TEXT NOT NULL,
    description TEXT NOT NULL,
    rules TEXT NOT NULL,
    active BOOLEAN NOT NULL DEFAULT TRUE,
    starts_at TIMESTAMPTZ NOT NULL,
    ends_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_programs_active_dates ON programs (active, starts_at, ends_at);
CREATE INDEX idx_programs_geo ON programs USING GIST ((ST_SetSRID(ST_MakePoint(lng, lat), 4326)::geography));

CREATE TABLE qr_codes (
    id TEXT PRIMARY KEY CHECK (id ~ '^qr_[A-Za-z0-9]{8}$'),
    program_id TEXT NOT NULL REFERENCES programs(id) ON DELETE CASCADE,
    merchant_id TEXT NOT NULL,
    code_hash TEXT NOT NULL UNIQUE,
    raw_payload TEXT NOT NULL,
    generated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    expires_at TIMESTAMPTZ NOT NULL,
    used BOOLEAN NOT NULL DEFAULT FALSE,
    used_by_user_id TEXT REFERENCES users(id),
    used_at TIMESTAMPTZ
);

CREATE INDEX idx_qr_codes_program ON qr_codes (program_id);
CREATE INDEX idx_qr_codes_expiration ON qr_codes (expires_at);

CREATE TABLE stamps (
    id TEXT PRIMARY KEY CHECK (id ~ '^selo_[A-Za-z0-9]{8}$'),
    user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    program_id TEXT NOT NULL REFERENCES programs(id) ON DELETE CASCADE,
    qr_code_id TEXT NOT NULL REFERENCES qr_codes(id) ON DELETE RESTRICT,
    qr_code_hash TEXT NOT NULL UNIQUE,
    acquired_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    validation_key TEXT NOT NULL UNIQUE,
    sequence INTEGER NOT NULL CHECK (sequence > 0),
    validated BOOLEAN NOT NULL DEFAULT TRUE,
    UNIQUE (user_id, program_id, sequence)
);

CREATE INDEX idx_stamps_user_program ON stamps (user_id, program_id);
CREATE INDEX idx_stamps_program ON stamps (program_id);
CREATE INDEX idx_stamps_acquired_at ON stamps (acquired_at DESC);
CREATE UNIQUE INDEX idx_stamps_user_program_day ON stamps (user_id, program_id, ((acquired_at AT TIME ZONE 'UTC')::date));

CREATE TABLE reward_redemptions (
    id TEXT PRIMARY KEY CHECK (id ~ '^rew_[A-Za-z0-9]{8}$'),
    user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    program_id TEXT NOT NULL REFERENCES programs(id) ON DELETE CASCADE,
    reward_code TEXT NOT NULL UNIQUE,
    completed_at TIMESTAMPTZ NOT NULL,
    expires_at TIMESTAMPTZ,
    redeemed BOOLEAN NOT NULL DEFAULT FALSE,
    redeemed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (user_id, program_id)
);

CREATE INDEX idx_reward_redemptions_user ON reward_redemptions (user_id, redeemed);
