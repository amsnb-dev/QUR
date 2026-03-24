-- 01_schema.sql â€” Auth tables + RLS
-- Idempotent: safe to run multiple times.

-- â”€â”€ Trigger helper â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€
CREATE OR REPLACE FUNCTION set_updated_at()
RETURNS TRIGGER AS $$
BEGIN
  NEW.updated_at = NOW();
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- â”€â”€ schools â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€
CREATE TABLE IF NOT EXISTS schools (
  id                  UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  name                TEXT        NOT NULL,
  city                TEXT,
  country             TEXT        NOT NULL DEFAULT 'SA',
  plan                TEXT        NOT NULL DEFAULT 'trial'
                                  CHECK (plan IN ('trial', 'basic', 'premium')),
  is_active           BOOLEAN     NOT NULL DEFAULT TRUE,
  default_monthly_fee NUMERIC(10, 2) NOT NULL DEFAULT 0,
  created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE OR REPLACE TRIGGER trg_schools_updated_at
  BEFORE UPDATE ON schools
  FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- â”€â”€ roles â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€
CREATE TABLE IF NOT EXISTS roles (
  id    SMALLINT PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
  name  TEXT NOT NULL UNIQUE,
  label TEXT NOT NULL
);

INSERT INTO roles (name, label) VALUES
  ('super_admin',  'Super Admin'),
  ('school_admin', 'Ù…Ø¯ÙŠØ± Ø§Ù„Ù…Ø¯Ø±Ø³Ø©'),
  ('supervisor',   'Ù…Ø´Ø±Ù ØªØ±Ø¨ÙˆÙŠ'),
  ('teacher',      'Ù…Ø¹Ù„Ù‘Ù…'),
  ('accountant',   'Ù…Ø­Ø§Ø³Ø¨')
ON CONFLICT (name) DO NOTHING;

-- â”€â”€ users â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€
CREATE TABLE IF NOT EXISTS users (
  id              UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  -- NULL = super_admin (platform-wide, no school affiliation)
  school_id       UUID        REFERENCES schools (id) ON DELETE RESTRICT,
  role_id         SMALLINT    NOT NULL REFERENCES roles (id),
  full_name       TEXT        NOT NULL,
  email           TEXT        NOT NULL,
  password_hash   TEXT        NOT NULL,
  phone           TEXT,
  is_active       BOOLEAN     NOT NULL DEFAULT TRUE,
  is_archived     BOOLEAN     NOT NULL DEFAULT FALSE,
  archived_at     TIMESTAMPTZ,
  failed_attempts SMALLINT    NOT NULL DEFAULT 0,
  locked_until    TIMESTAMPTZ,
  last_login_at   TIMESTAMPTZ,
  invite_token    TEXT,
  invite_expires  TIMESTAMPTZ,
  created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  CONSTRAINT chk_users_archive_consistency CHECK (
    (is_archived = FALSE AND archived_at IS NULL)
    OR (is_archived = TRUE AND archived_at IS NOT NULL)
  )
);

-- Unique email per school (NULLs treated as distinct by Postgres, so super_admin
-- entries with school_id IS NULL each get their own slot â€” which is correct).
CREATE UNIQUE INDEX IF NOT EXISTS uq_users_school_email
  ON users (school_id, email)
  WHERE school_id IS NOT NULL;

-- Separate unique index for super_admins (school_id IS NULL)
CREATE UNIQUE INDEX IF NOT EXISTS uq_users_super_email
  ON users (email)
  WHERE school_id IS NULL;

CREATE OR REPLACE TRIGGER trg_users_updated_at
  BEFORE UPDATE ON users
  FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- â”€â”€ refresh_tokens â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€
CREATE TABLE IF NOT EXISTS refresh_tokens (
  id         UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id    UUID        NOT NULL REFERENCES users (id) ON DELETE CASCADE,
  -- Only the SHA-256 hex digest is stored â€” never the raw token.
  token_hash TEXT        NOT NULL UNIQUE,
  expires_at TIMESTAMPTZ NOT NULL,
  revoked    BOOLEAN     NOT NULL DEFAULT FALSE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- â”€â”€ Indexes â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€
CREATE INDEX IF NOT EXISTS idx_users_email  ON users (email);
CREATE INDEX IF NOT EXISTS idx_users_school ON users (school_id);
CREATE INDEX IF NOT EXISTS idx_rt_user      ON refresh_tokens (user_id);
-- Partial index: only un-revoked tokens need fast lookup by expiry.
CREATE INDEX IF NOT EXISTS idx_rt_active_expires
  ON refresh_tokens (expires_at)
  WHERE revoked = FALSE;

-- â”€â”€ Row-Level Security â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€
ALTER TABLE users ENABLE ROW LEVEL SECURITY;

-- Policy:
--   Normal user  â†’ SET LOCAL app.school_id = '<uuid>'  â†’ sees own school only
--   Super Admin  â†’ SET LOCAL app.school_id = ''        â†’ sees everything
DROP POLICY IF EXISTS school_isolation ON users;
CREATE POLICY school_isolation ON users
  USING (
    school_id = NULLIF(current_setting('app.school_id', TRUE), '')::UUID
    OR COALESCE(current_setting('app.school_id', TRUE), '') = ''
  );

-- refresh_tokens: no school_id, no RLS needed â€” access controlled by user_id FK
-- and the fact that token_hash is only known to the token owner.
