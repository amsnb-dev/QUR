-- ============================================================
--  03_core_mvp.sql
--  Ø§Ù„Ø¬Ø¯Ø§ÙˆÙ„ Ø§Ù„ØªØ´ØºÙŠÙ„ÙŠØ© Ø§Ù„Ø£Ø³Ø§Ø³ÙŠØ© Ù„Ù„Ù€ MVP
--  Multi-Tenant Â· RLS Â· Ø£Ø±Ø´ÙØ© Ø¨Ø¯Ù„ Ø­Ø°Ù Â· FK Ù…Ø±ÙƒÙ‘Ø¨
--  Idempotent: Ø¢Ù…Ù† Ù„Ù„ØªØ´ØºÙŠÙ„ Ø£ÙƒØ«Ø± Ù…Ù† Ù…Ø±Ø©
-- ============================================================
--  Ø§Ù„Ù…ØªØ·Ù„Ø¨Ø§Øª Ø§Ù„Ù…Ø³Ø¨Ù‚Ø©: 00_init.sql + 01_schema.sql
--  (schools Â· roles Â· users Â· refresh_tokens Â· set_updated_at)
-- ============================================================

BEGIN;

-- â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•
--  1. teachers
-- â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•

CREATE TABLE IF NOT EXISTS teachers (
    id               UUID          PRIMARY KEY DEFAULT gen_random_uuid(),
    school_id        UUID          NOT NULL REFERENCES schools (id) ON DELETE RESTRICT,
    user_id          UUID          NOT NULL REFERENCES users (id)   ON DELETE RESTRICT,
    specialization   TEXT          CHECK (specialization IN ('hafz', 'tajweed', 'both')),
    qualification    TEXT,
    hire_date        DATE,
    base_salary      NUMERIC(10,2) NOT NULL DEFAULT 0,
    housing_allow    NUMERIC(10,2) NOT NULL DEFAULT 0,
    transport_allow  NUMERIC(10,2) NOT NULL DEFAULT 0,
    -- IBAN Ù…Ø´ÙÙ‘Ø± Ø¹Ù„Ù‰ Ù…Ø³ØªÙˆÙ‰ Ø§Ù„ØªØ·Ø¨ÙŠÙ‚ Ù‚Ø¨Ù„ Ø§Ù„ØªØ®Ø²ÙŠÙ†
    bank_iban_enc    TEXT,
    is_active        BOOLEAN       NOT NULL DEFAULT TRUE,
    -- Ø£Ø±Ø´ÙØ© Ø¨Ø¯Ù„ Ø­Ø°Ù
    is_archived      BOOLEAN       NOT NULL DEFAULT FALSE,
    archived_at      TIMESTAMPTZ,
    created_at       TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    -- teacher Ù…Ø±ØªØ¨Ø· Ø¨Ù€ user ÙˆØ§Ø­Ø¯ Ø¯Ø§Ø®Ù„ Ù†ÙØ³ Ø§Ù„Ù…Ø¯Ø±Ø³Ø©
    CONSTRAINT uq_teachers_school_user UNIQUE (school_id, user_id),
    -- UNIQUE (id, school_id) Ø´Ø±Ø· Ù…Ø³Ø¨Ù‚ Ù„Ù„Ù€ FK Ø§Ù„Ù…Ø±ÙƒÙ‘Ø¨
    CONSTRAINT uq_teachers_id_school   UNIQUE (id, school_id),
    CONSTRAINT chk_teachers_archive    CHECK (
        (is_archived = FALSE AND archived_at IS NULL)
        OR (is_archived = TRUE  AND archived_at IS NOT NULL)
    )
);

CREATE OR REPLACE TRIGGER trg_teachers_updated_at
    BEFORE UPDATE ON teachers
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- â”€â”€ indexes â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€
CREATE INDEX IF NOT EXISTS idx_teachers_school
    ON teachers (school_id);
CREATE INDEX IF NOT EXISTS idx_teachers_user
    ON teachers (user_id);
-- Partial: Ø§Ù„Ø§Ø³ØªØ¹Ù„Ø§Ù…Ø§Øª Ø§Ù„Ø¹Ø§Ø¯ÙŠØ© ØªØ³ØªÙ‡Ø¯Ù Ø§Ù„Ù†Ø´Ø·ÙŠÙ† ÙÙ‚Ø·
CREATE INDEX IF NOT EXISTS idx_teachers_active
    ON teachers (school_id)
    WHERE is_archived = FALSE;

-- â”€â”€ RLS â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€
ALTER TABLE teachers ENABLE ROW LEVEL SECURITY;

DROP POLICY IF EXISTS school_isolation ON teachers;
CREATE POLICY school_isolation ON teachers
    USING (
        school_id = NULLIF(current_setting('app.school_id', TRUE), '')::UUID
        OR COALESCE(current_setting('app.school_id', TRUE), '') = ''
    );

-- â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•
--  2. groups (Ø§Ù„Ø­Ù„Ù‚Ø§Øª)
-- â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•

CREATE TABLE IF NOT EXISTS groups (
    id          UUID       PRIMARY KEY DEFAULT gen_random_uuid(),
    school_id   UUID       NOT NULL REFERENCES schools  (id)   ON DELETE RESTRICT,
    teacher_id  UUID       NOT NULL REFERENCES teachers (id)   ON DELETE RESTRICT,
    name        TEXT       NOT NULL,
    level       TEXT       CHECK (level IN ('beginner', 'intermediate', 'advanced', 'mixed')),
    room        TEXT,
    capacity    SMALLINT   NOT NULL DEFAULT 30,
    -- Ø£ÙŠØ§Ù… Ø§Ù„Ø£Ø³Ø¨ÙˆØ¹: 0=Ø£Ø­Ø¯ â€¦ 6=Ø³Ø¨Øª
    days        SMALLINT[] NOT NULL DEFAULT '{}',
    start_time  TIME,
    end_time    TIME,
    is_active   BOOLEAN    NOT NULL DEFAULT TRUE,
    is_archived BOOLEAN    NOT NULL DEFAULT FALSE,
    archived_at TIMESTAMPTZ,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_groups_school_name  UNIQUE (school_id, name),
    CONSTRAINT uq_groups_id_school    UNIQUE (id, school_id),
    CONSTRAINT chk_groups_archive     CHECK (
        (is_archived = FALSE AND archived_at IS NULL)
        OR (is_archived = TRUE  AND archived_at IS NOT NULL)
    ),
    CONSTRAINT chk_groups_time        CHECK (end_time IS NULL OR end_time > start_time)
);

CREATE OR REPLACE TRIGGER trg_groups_updated_at
    BEFORE UPDATE ON groups
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- â”€â”€ indexes â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€
CREATE INDEX IF NOT EXISTS idx_groups_school
    ON groups (school_id);
CREATE INDEX IF NOT EXISTS idx_groups_teacher
    ON groups (teacher_id);
CREATE INDEX IF NOT EXISTS idx_groups_active
    ON groups (school_id)
    WHERE is_archived = FALSE;

-- â”€â”€ RLS â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€
ALTER TABLE groups ENABLE ROW LEVEL SECURITY;

DROP POLICY IF EXISTS school_isolation ON groups;
CREATE POLICY school_isolation ON groups
    USING (
        school_id = NULLIF(current_setting('app.school_id', TRUE), '')::UUID
        OR COALESCE(current_setting('app.school_id', TRUE), '') = ''
    );

-- â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•
--  3. students
-- â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•

CREATE TABLE IF NOT EXISTS students (
    id               UUID          PRIMARY KEY DEFAULT gen_random_uuid(),
    school_id        UUID          NOT NULL REFERENCES schools (id) ON DELETE RESTRICT,
    -- Ù„Ø§ ÙŠÙˆØ¬Ø¯ group_id Ù…Ø¨Ø§Ø´Ø± â€” Ø§Ù„Ø¹Ù„Ø§Ù‚Ø© Ø¹Ø¨Ø± student_groups
    full_name        TEXT          NOT NULL,
    date_of_birth    DATE,
    -- Ù…Ø´ÙÙ‘Ø± Ø¹Ù„Ù‰ Ù…Ø³ØªÙˆÙ‰ Ø§Ù„ØªØ·Ø¨ÙŠÙ‚
    national_id      TEXT,
    guardian_name    TEXT,
    -- Ù…Ø´ÙÙ‘Ø± Ø¹Ù„Ù‰ Ù…Ø³ØªÙˆÙ‰ Ø§Ù„ØªØ·Ø¨ÙŠÙ‚
    guardian_phone   TEXT,
    guardian_phone2  TEXT,
    enrollment_date  DATE          NOT NULL DEFAULT CURRENT_DATE,
    -- 0.0â€“30.0 Ø¬Ø²Ø¡
    memorized_parts  NUMERIC(4,1)  NOT NULL DEFAULT 0
                     CHECK (memorized_parts BETWEEN 0 AND 30),
    level_on_entry   TEXT          CHECK (level_on_entry IN ('beginner', 'intermediate', 'advanced')),
    monthly_fee      NUMERIC(10,2),
    fee_exemption    TEXT          NOT NULL DEFAULT 'none'
                     CHECK (fee_exemption IN ('none', 'partial', 'full')),
    status           TEXT          NOT NULL DEFAULT 'active'
                     CHECK (status IN ('active', 'inactive', 'graduated', 'transferred')),
    notes            TEXT,
    is_archived      BOOLEAN       NOT NULL DEFAULT FALSE,
    archived_at      TIMESTAMPTZ,
    created_at       TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_students_id_school UNIQUE (id, school_id),
    CONSTRAINT chk_students_archive  CHECK (
        (is_archived = FALSE AND archived_at IS NULL)
        OR (is_archived = TRUE  AND archived_at IS NOT NULL)
    )
);

CREATE OR REPLACE TRIGGER trg_students_updated_at
    BEFORE UPDATE ON students
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- â”€â”€ indexes â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€
CREATE INDEX IF NOT EXISTS idx_students_school
    ON students (school_id);
CREATE INDEX IF NOT EXISTS idx_students_status
    ON students (school_id, status);
CREATE INDEX IF NOT EXISTS idx_students_name
    ON students (school_id, full_name);
CREATE INDEX IF NOT EXISTS idx_students_active
    ON students (school_id, status)
    WHERE is_archived = FALSE;

-- â”€â”€ RLS â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€
ALTER TABLE students ENABLE ROW LEVEL SECURITY;

DROP POLICY IF EXISTS school_isolation ON students;
CREATE POLICY school_isolation ON students
    USING (
        school_id = NULLIF(current_setting('app.school_id', TRUE), '')::UUID
        OR COALESCE(current_setting('app.school_id', TRUE), '') = ''
    );

-- â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•
--  4. student_groups  (Ø·Ø§Ù„Ø¨ â† â†’ Ø­Ù„Ù‚Ø©ØŒ Ø¹Ù„Ø§Ù‚Ø© Ù…ØªØ¹Ø¯Ø¯Ø© Ù…Ø¹ ØªÙˆØ§Ø±ÙŠØ®)
-- â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•

CREATE TABLE IF NOT EXISTS student_groups (
    id         UUID    PRIMARY KEY DEFAULT gen_random_uuid(),
    school_id  UUID    NOT NULL REFERENCES schools (id) ON DELETE RESTRICT,
    -- FK Ù…Ø±ÙƒÙ‘Ø¨ ÙŠØ¶Ù…Ù† Ø£Ù† Ø§Ù„Ø·Ø§Ù„Ø¨ ÙˆØ§Ù„Ø­Ù„Ù‚Ø© Ù…Ù† Ù†ÙØ³ Ø§Ù„Ù…Ø¯Ø±Ø³Ø©
    student_id UUID    NOT NULL,
    group_id   UUID    NOT NULL,
    start_date DATE    NOT NULL DEFAULT CURRENT_DATE,
    -- NULL = Ù„Ø§ ÙŠØ²Ø§Ù„ ÙÙŠ Ø§Ù„Ø­Ù„Ù‚Ø©
    end_date   DATE,
    is_primary BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT fk_sg_student_school
        FOREIGN KEY (student_id, school_id)
        REFERENCES students (id, school_id) ON DELETE RESTRICT,
    CONSTRAINT fk_sg_group_school
        FOREIGN KEY (group_id, school_id)
        REFERENCES groups (id, school_id) ON DELETE RESTRICT,
    CONSTRAINT chk_sg_dates
        CHECK (end_date IS NULL OR end_date > start_date)
);

-- Ø·Ø§Ù„Ø¨ ÙˆØ§Ø­Ø¯ ÙÙ‚Ø· Ù„Ù‡ Ø­Ù„Ù‚Ø© Ø±Ø¦ÙŠØ³ÙŠØ© Ù…ÙØªÙˆØ­Ø© ÙÙŠ Ù†ÙØ³ Ø§Ù„ÙˆÙ‚Øª
CREATE UNIQUE INDEX IF NOT EXISTS uq_sg_one_primary_active
    ON student_groups (school_id, student_id)
    WHERE is_primary = TRUE AND end_date IS NULL;

-- Ù„Ø§ ØªÙƒØ±Ø§Ø± Ù„Ù†ÙØ³ Ø§Ù„Ø·Ø§Ù„Ø¨ ÙÙŠ Ù†ÙØ³ Ø§Ù„Ø­Ù„Ù‚Ø© Ø§Ù„Ù…ÙØªÙˆØ­Ø©
CREATE UNIQUE INDEX IF NOT EXISTS uq_sg_no_duplicate_active
    ON student_groups (school_id, student_id, group_id)
    WHERE end_date IS NULL;

-- â”€â”€ indexes â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€
CREATE INDEX IF NOT EXISTS idx_sg_student
    ON student_groups (student_id, end_date NULLS FIRST);
CREATE INDEX IF NOT EXISTS idx_sg_group
    ON student_groups (group_id, end_date NULLS FIRST);
CREATE INDEX IF NOT EXISTS idx_sg_school
    ON student_groups (school_id, is_primary, end_date NULLS FIRST);

-- â”€â”€ RLS â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€
ALTER TABLE student_groups ENABLE ROW LEVEL SECURITY;

DROP POLICY IF EXISTS school_isolation ON student_groups;
CREATE POLICY school_isolation ON student_groups
    USING (
        school_id = NULLIF(current_setting('app.school_id', TRUE), '')::UUID
        OR COALESCE(current_setting('app.school_id', TRUE), '') = ''
    );

-- â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•
--  5. Ø¬Ø¯ÙˆÙ„ Ù…Ø±Ø¬Ø¹ÙŠ: surahs (Ø«Ø§Ø¨ØªØŒ Ù„Ø§ school_idØŒ Ù„Ø§ RLS)
-- â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•

CREATE TABLE IF NOT EXISTS surahs (
    number      SMALLINT PRIMARY KEY CHECK (number BETWEEN 1 AND 114),
    name_ar     TEXT     NOT NULL,
    name_en     TEXT     NOT NULL,
    juz_start   SMALLINT NOT NULL CHECK (juz_start BETWEEN 1 AND 30),
    juz_end     SMALLINT NOT NULL CHECK (juz_end   BETWEEN 1 AND 30),
    verse_count SMALLINT NOT NULL
);

-- â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•
--  6. attendance
-- â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•

CREATE TABLE IF NOT EXISTS attendance (
    id          UUID  PRIMARY KEY DEFAULT gen_random_uuid(),
    school_id   UUID  NOT NULL REFERENCES schools  (id) ON DELETE RESTRICT,
    student_id  UUID  NOT NULL REFERENCES students (id) ON DELETE RESTRICT,
    -- group_id nullable: Ø¬Ù„Ø³Ø© Ø®Ø§Ø±Ø¬ Ø§Ù„Ø­Ù„Ù‚Ø© Ù…Ø³Ù…ÙˆØ­ Ø¨Ù‡Ø§
    group_id    UUID,
    recorded_by UUID  NOT NULL REFERENCES users    (id),
    date        DATE  NOT NULL,
    status      TEXT  NOT NULL
                CHECK (status IN ('present', 'absent', 'excused', 'late')),
    note        TEXT,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    -- Ø³Ø¬Ù„ ÙˆØ§Ø­Ø¯ Ù„ÙƒÙ„ Ø·Ø§Ù„Ø¨ ÙÙŠ ÙƒÙ„ ÙŠÙˆÙ… Ù„ÙƒÙ„ Ø­Ù„Ù‚Ø©
    CONSTRAINT uq_attendance_student_group_date
        UNIQUE (school_id, student_id, group_id, date),
    -- FK Ù…Ø±ÙƒÙ‘Ø¨: group_id + school_id â†’ groups
    CONSTRAINT fk_att_group_school
        FOREIGN KEY (group_id, school_id)
        REFERENCES groups (id, school_id) ON DELETE SET NULL
);

-- â”€â”€ indexes â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€
CREATE INDEX IF NOT EXISTS idx_att_school_date
    ON attendance (school_id, date DESC);
CREATE INDEX IF NOT EXISTS idx_att_student
    ON attendance (student_id, date DESC);
CREATE INDEX IF NOT EXISTS idx_att_group_date
    ON attendance (group_id, date DESC)
    WHERE group_id IS NOT NULL;

-- â”€â”€ RLS â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€
ALTER TABLE attendance ENABLE ROW LEVEL SECURITY;

DROP POLICY IF EXISTS school_isolation ON attendance;
CREATE POLICY school_isolation ON attendance
    USING (
        school_id = NULLIF(current_setting('app.school_id', TRUE), '')::UUID
        OR COALESCE(current_setting('app.school_id', TRUE), '') = ''
    );

-- â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•
--  7. memorization_logs
-- â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•

CREATE TABLE IF NOT EXISTS memorization_logs (
    id           UUID     PRIMARY KEY DEFAULT gen_random_uuid(),
    school_id    UUID     NOT NULL REFERENCES schools  (id)      ON DELETE RESTRICT,
    student_id   UUID     NOT NULL REFERENCES students (id)      ON DELETE RESTRICT,
    group_id     UUID,
    recorded_by  UUID     NOT NULL REFERENCES users    (id),
    date         DATE     NOT NULL DEFAULT CURRENT_DATE,
    surah_number SMALLINT NOT NULL REFERENCES surahs   (number),
    from_verse   SMALLINT NOT NULL CHECK (from_verse >= 1),
    to_verse     SMALLINT NOT NULL CHECK (to_verse >= from_verse),
    juz          SMALLINT CHECK (juz BETWEEN 1 AND 30),
    entry_type   TEXT     NOT NULL DEFAULT 'new'
                 CHECK (entry_type IN ('new', 'review')),
    grade        SMALLINT NOT NULL CHECK (grade BETWEEN 1 AND 5),
    tajweed_errors TEXT[] NOT NULL DEFAULT '{}',
    notes        TEXT,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT fk_mem_group_school
        FOREIGN KEY (group_id, school_id)
        REFERENCES groups (id, school_id) ON DELETE SET NULL
);

-- â”€â”€ indexes â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€
CREATE INDEX IF NOT EXISTS idx_mem_student
    ON memorization_logs (student_id, date DESC);
CREATE INDEX IF NOT EXISTS idx_mem_school_date
    ON memorization_logs (school_id, date DESC);
CREATE INDEX IF NOT EXISTS idx_mem_type
    ON memorization_logs (student_id, entry_type);
CREATE INDEX IF NOT EXISTS idx_mem_group
    ON memorization_logs (group_id, date DESC)
    WHERE group_id IS NOT NULL;

-- â”€â”€ RLS â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€
ALTER TABLE memorization_logs ENABLE ROW LEVEL SECURITY;

DROP POLICY IF EXISTS school_isolation ON memorization_logs;
CREATE POLICY school_isolation ON memorization_logs
    USING (
        school_id = NULLIF(current_setting('app.school_id', TRUE), '')::UUID
        OR COALESCE(current_setting('app.school_id', TRUE), '') = ''
    );

-- â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•
--  8. recitation_sessions
-- â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•

CREATE TABLE IF NOT EXISTS recitation_sessions (
    id            UUID     PRIMARY KEY DEFAULT gen_random_uuid(),
    school_id     UUID     NOT NULL REFERENCES schools  (id) ON DELETE RESTRICT,
    student_id    UUID     NOT NULL REFERENCES students (id) ON DELETE RESTRICT,
    group_id      UUID,
    recorded_by   UUID     NOT NULL REFERENCES users    (id),
    date          DATE     NOT NULL DEFAULT CURRENT_DATE,
    session_type  TEXT     NOT NULL
                  CHECK (session_type IN ('hafz', 'review', 'tajweed', 'exam')),
    surah_number  SMALLINT REFERENCES surahs (number),
    pages_range   TEXT,
    hafz_score    NUMERIC(4,1) CHECK (hafz_score    BETWEEN 0 AND 10),
    tajweed_score NUMERIC(4,1) CHECK (tajweed_score BETWEEN 0 AND 10),
    tajweed_errors TEXT[]  NOT NULL DEFAULT '{}',
    notes         TEXT,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT fk_rec_group_school
        FOREIGN KEY (group_id, school_id)
        REFERENCES groups (id, school_id) ON DELETE SET NULL
);

-- â”€â”€ indexes â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€
CREATE INDEX IF NOT EXISTS idx_rec_student
    ON recitation_sessions (student_id, date DESC);
CREATE INDEX IF NOT EXISTS idx_rec_school_date
    ON recitation_sessions (school_id, date DESC);
CREATE INDEX IF NOT EXISTS idx_rec_group
    ON recitation_sessions (group_id, date DESC)
    WHERE group_id IS NOT NULL;

-- â”€â”€ RLS â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€
ALTER TABLE recitation_sessions ENABLE ROW LEVEL SECURITY;

DROP POLICY IF EXISTS school_isolation ON recitation_sessions;
CREATE POLICY school_isolation ON recitation_sessions
    USING (
        school_id = NULLIF(current_setting('app.school_id', TRUE), '')::UUID
        OR COALESCE(current_setting('app.school_id', TRUE), '') = ''
    );

-- â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•
--  9. invoices
-- â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•

CREATE TABLE IF NOT EXISTS invoices (
    id            UUID          PRIMARY KEY DEFAULT gen_random_uuid(),
    school_id     UUID          NOT NULL REFERENCES schools  (id) ON DELETE RESTRICT,
    student_id    UUID          NOT NULL REFERENCES students (id) ON DELETE RESTRICT,
    -- Ø£ÙˆÙ„ ÙŠÙˆÙ… Ù…Ù† Ø§Ù„Ø´Ù‡Ø± Ø¯Ø§Ø¦Ù…Ø§Ù‹ (2025-03-01)
    billing_month DATE          NOT NULL,
    amount        NUMERIC(10,2) NOT NULL CHECK (amount >= 0),
    status        TEXT          NOT NULL DEFAULT 'unpaid'
                  CHECK (status IN ('unpaid', 'paid', 'partial', 'waived')),
    due_date      DATE          NOT NULL,
    notes         TEXT,
    created_by    UUID          NOT NULL REFERENCES users (id),
    created_at    TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    -- ÙØ§ØªÙˆØ±Ø© ÙˆØ§Ø­Ø¯Ø© Ù„ÙƒÙ„ Ø·Ø§Ù„Ø¨ ÙÙŠ ÙƒÙ„ Ø´Ù‡Ø±
    CONSTRAINT uq_invoices_student_month
        UNIQUE (school_id, student_id, billing_month),
    CONSTRAINT chk_billing_first_day
        CHECK (EXTRACT(DAY FROM billing_month) = 1)
);

CREATE OR REPLACE TRIGGER trg_invoices_updated_at
    BEFORE UPDATE ON invoices
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- â”€â”€ indexes â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€
CREATE INDEX IF NOT EXISTS idx_inv_school
    ON invoices (school_id, billing_month DESC);
CREATE INDEX IF NOT EXISTS idx_inv_student
    ON invoices (student_id);
CREATE INDEX IF NOT EXISTS idx_inv_status
    ON invoices (school_id, status);

-- â”€â”€ RLS â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€
ALTER TABLE invoices ENABLE ROW LEVEL SECURITY;

DROP POLICY IF EXISTS school_isolation ON invoices;
CREATE POLICY school_isolation ON invoices
    USING (
        school_id = NULLIF(current_setting('app.school_id', TRUE), '')::UUID
        OR COALESCE(current_setting('app.school_id', TRUE), '') = ''
    );

-- â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•
--  10. payments
-- â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•

CREATE TABLE IF NOT EXISTS payments (
    id             UUID          PRIMARY KEY DEFAULT gen_random_uuid(),
    school_id      UUID          NOT NULL REFERENCES schools  (id) ON DELETE RESTRICT,
    invoice_id     UUID          NOT NULL REFERENCES invoices (id) ON DELETE RESTRICT,
    student_id     UUID          NOT NULL REFERENCES students (id) ON DELETE RESTRICT,
    amount         NUMERIC(10,2) NOT NULL CHECK (amount > 0),
    payment_method TEXT          NOT NULL
                   CHECK (payment_method IN ('cash', 'transfer', 'card', 'app')),
    payment_date   DATE          NOT NULL DEFAULT CURRENT_DATE,
    -- Ø±Ù‚Ù… Ø§Ù„Ø¥ÙŠØµØ§Ù„: ÙŠÙÙ…Ù„Ø£ ØªÙ„Ù‚Ø§Ø¦ÙŠØ§Ù‹ Ø¨Ù€ trigger Ø¹Ù†Ø¯ Ø§Ù„Ø¥Ø¯Ø±Ø§Ø¬
    receipt_number TEXT,
    notes          TEXT,
    recorded_by    UUID          NOT NULL REFERENCES users (id),
    created_at     TIMESTAMPTZ   NOT NULL DEFAULT NOW()
);

-- â”€â”€ indexes â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€
CREATE INDEX IF NOT EXISTS idx_pay_school_date
    ON payments (school_id, payment_date DESC);
CREATE INDEX IF NOT EXISTS idx_pay_invoice
    ON payments (invoice_id);
CREATE INDEX IF NOT EXISTS idx_pay_student
    ON payments (student_id);

-- â”€â”€ RLS â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€
ALTER TABLE payments ENABLE ROW LEVEL SECURITY;

DROP POLICY IF EXISTS school_isolation ON payments;
CREATE POLICY school_isolation ON payments
    USING (
        school_id = NULLIF(current_setting('app.school_id', TRUE), '')::UUID
        OR COALESCE(current_setting('app.school_id', TRUE), '') = ''
    );

-- â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•
--  11. expenses
-- â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•

CREATE TABLE IF NOT EXISTS expenses (
    id             UUID          PRIMARY KEY DEFAULT gen_random_uuid(),
    school_id      UUID          NOT NULL REFERENCES schools (id) ON DELETE RESTRICT,
    category       TEXT          NOT NULL
                   CHECK (category IN ('utilities', 'supplies', 'maintenance',
                                       'books', 'events', 'other')),
    amount         NUMERIC(10,2) NOT NULL CHECK (amount > 0),
    description    TEXT,
    expense_date   DATE          NOT NULL DEFAULT CURRENT_DATE,
    payment_method TEXT          CHECK (payment_method IN ('cash', 'transfer', 'card')),
    recorded_by    UUID          NOT NULL REFERENCES users (id),
    created_at     TIMESTAMPTZ   NOT NULL DEFAULT NOW()
);

-- â”€â”€ indexes â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€
CREATE INDEX IF NOT EXISTS idx_exp_school_date
    ON expenses (school_id, expense_date DESC);
CREATE INDEX IF NOT EXISTS idx_exp_category
    ON expenses (school_id, category);

-- â”€â”€ RLS â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€
ALTER TABLE expenses ENABLE ROW LEVEL SECURITY;

DROP POLICY IF EXISTS school_isolation ON expenses;
CREATE POLICY school_isolation ON expenses
    USING (
        school_id = NULLIF(current_setting('app.school_id', TRUE), '')::UUID
        OR COALESCE(current_setting('app.school_id', TRUE), '') = ''
    );

-- â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•
--  12. payroll_records
-- â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•

CREATE TABLE IF NOT EXISTS payroll_records (
    id                UUID          PRIMARY KEY DEFAULT gen_random_uuid(),
    school_id         UUID          NOT NULL REFERENCES schools  (id) ON DELETE RESTRICT,
    teacher_id        UUID          NOT NULL REFERENCES teachers (id) ON DELETE RESTRICT,
    -- Ø£ÙˆÙ„ ÙŠÙˆÙ… Ù…Ù† Ø§Ù„Ø´Ù‡Ø± Ø¯Ø§Ø¦Ù…Ø§Ù‹
    payroll_month     DATE          NOT NULL,
    base_salary       NUMERIC(10,2) NOT NULL CHECK (base_salary >= 0),
    housing_allow     NUMERIC(10,2) NOT NULL DEFAULT 0,
    transport_allow   NUMERIC(10,2) NOT NULL DEFAULT 0,
    bonus             NUMERIC(10,2) NOT NULL DEFAULT 0,
    absence_deduction NUMERIC(10,2) NOT NULL DEFAULT 0,
    gosi_deduction    NUMERIC(10,2) NOT NULL DEFAULT 0,
    other_deduction   NUMERIC(10,2) NOT NULL DEFAULT 0,
    -- ØµØ§ÙÙŠ Ø§Ù„Ø±Ø§ØªØ¨: ÙŠÙØ­Ø³Ø¨ ÙˆÙŠÙØ®Ø²ÙŽÙ‘Ù† ØªÙ„Ù‚Ø§Ø¦ÙŠØ§Ù‹ Ø¨Ù€ trigger
    net_salary        NUMERIC(10,2),
    status            TEXT          NOT NULL DEFAULT 'draft'
                      CHECK (status IN ('draft', 'paid')),
    paid_date         DATE,
    processed_by      UUID          NOT NULL REFERENCES users (id),
    created_at        TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_payroll_teacher_month
        UNIQUE (school_id, teacher_id, payroll_month),
    CONSTRAINT chk_payroll_first_day
        CHECK (EXTRACT(DAY FROM payroll_month) = 1),
    CONSTRAINT chk_payroll_paid_date
        CHECK (paid_date IS NULL OR paid_date >= payroll_month)
);

CREATE OR REPLACE TRIGGER trg_payroll_updated_at
    BEFORE UPDATE ON payroll_records
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- â”€â”€ indexes â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€
CREATE INDEX IF NOT EXISTS idx_pr_school_month
    ON payroll_records (school_id, payroll_month DESC);
CREATE INDEX IF NOT EXISTS idx_pr_teacher
    ON payroll_records (teacher_id);

-- â”€â”€ RLS â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€
ALTER TABLE payroll_records ENABLE ROW LEVEL SECURITY;

DROP POLICY IF EXISTS school_isolation ON payroll_records;
CREATE POLICY school_isolation ON payroll_records
    USING (
        school_id = NULLIF(current_setting('app.school_id', TRUE), '')::UUID
        OR COALESCE(current_setting('app.school_id', TRUE), '') = ''
    );

-- â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•
--  13. audit_logs  (INSERT ONLY â€” Ù„Ø§ UPDATE Â· Ù„Ø§ DELETE)
-- â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•

CREATE TABLE IF NOT EXISTS audit_logs (
    id         BIGINT      PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    school_id  UUID        REFERENCES schools (id),  -- NULL Ù„Ù„Ø¹Ù…Ù„ÙŠØ§Øª Ø§Ù„Ø¹Ø§Ù…Ø©
    user_id    UUID        REFERENCES users   (id),  -- NULL Ù„Ù„Ù†Ø¸Ø§Ù…
    action     TEXT        NOT NULL,   -- 'INSERT' | 'UPDATE' | 'DELETE' | 'ARCHIVE' | 'LOGIN'
    table_name TEXT        NOT NULL,
    record_id  TEXT,
    old_values JSONB,
    new_values JSONB,
    ip_address INET,
    user_agent TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Trigger: ÙŠÙ…Ù†Ø¹ UPDATE Ø¹Ù„Ù‰ audit_logs
CREATE OR REPLACE FUNCTION prevent_audit_mutation()
RETURNS TRIGGER AS $$
BEGIN
    RAISE EXCEPTION
        'audit_logs Ù‡Ùˆ Ø³Ø¬Ù„ Ù„Ù„Ù‚Ø±Ø§Ø¡Ø© ÙÙ‚Ø· â€” % Ù…Ø±ÙÙˆØ¶ Ø¹Ù„Ù‰ Ø§Ù„Ø³Ø¬Ù„ id=%',
        TG_OP, OLD.id;
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

-- DROP Ø«Ù… CREATE Ù„Ø¶Ù…Ø§Ù† idempotency
DROP TRIGGER IF EXISTS trg_audit_no_update ON audit_logs;
CREATE TRIGGER trg_audit_no_update
    BEFORE UPDATE ON audit_logs
    FOR EACH ROW EXECUTE FUNCTION prevent_audit_mutation();

DROP TRIGGER IF EXISTS trg_audit_no_delete ON audit_logs;
CREATE TRIGGER trg_audit_no_delete
    BEFORE DELETE ON audit_logs
    FOR EACH ROW EXECUTE FUNCTION prevent_audit_mutation();

-- â”€â”€ indexes â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€
CREATE INDEX IF NOT EXISTS idx_audit_school
    ON audit_logs (school_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_audit_user
    ON audit_logs (user_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_audit_table
    ON audit_logs (table_name, created_at DESC);

-- audit_logs Ù„Ø§ ØªØ­ØªØ§Ø¬ RLS â€” Ø§Ù„ÙˆØµÙˆÙ„ ÙŠÙØ¯Ø§Ø± Ø¹Ù„Ù‰ Ù…Ø³ØªÙˆÙ‰ Ø§Ù„ØªØ·Ø¨ÙŠÙ‚
-- (super_admin ÙŠØ±Ù‰ Ø§Ù„ÙƒÙ„ØŒ school_admin ÙŠØ±Ù‰ Ù…Ø¯Ø±Ø³ØªÙ‡ ÙÙ‚Ø· Ø¹Ø¨Ø± API)

-- â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•
--  14. Ø¯ÙˆØ§Ù„ Ø§Ù„Ø£Ø±Ø´ÙØ© Ø§Ù„Ø¢Ù…Ù†Ø©
--      ØªØªØ­Ù‚Ù‚ Ù…Ù† Ø§Ù„Ø´Ø±ÙˆØ· Ù‚Ø¨Ù„ Ø§Ù„Ø³Ù…Ø§Ø­ Ø¨Ø§Ù„Ø£Ø±Ø´ÙØ©
-- â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•

-- Ø£Ø±Ø´ÙØ© Ø·Ø§Ù„Ø¨: Ù„Ø§ Ø­Ù„Ù‚Ø§Øª Ù…ÙØªÙˆØ­Ø© + Ù„Ø§ ÙÙˆØ§ØªÙŠØ± ØºÙŠØ± Ù…Ø³Ø¯Ø¯Ø©
CREATE OR REPLACE FUNCTION archive_student(
    p_student_id  UUID,
    p_school_id   UUID,
    p_archived_by UUID
) RETURNS VOID AS $$
DECLARE
    v_open_groups INT;
    v_unpaid_inv  INT;
BEGIN
    SELECT COUNT(*) INTO v_open_groups
    FROM   student_groups
    WHERE  student_id = p_student_id
      AND  school_id  = p_school_id
      AND  end_date IS NULL;

    IF v_open_groups > 0 THEN
        RAISE EXCEPTION
            'Ù„Ø§ ÙŠÙ…ÙƒÙ† Ø£Ø±Ø´ÙØ© Ø§Ù„Ø·Ø§Ù„Ø¨: Ù„Ø¯ÙŠÙ‡ % Ø­Ù„Ù‚Ø©/Ø­Ù„Ù‚Ø§Øª Ù…ÙØªÙˆØ­Ø© â€” Ø£ØºÙ„Ù‚Ù‡Ø§ Ø£ÙˆÙ„Ø§Ù‹.',
            v_open_groups;
    END IF;

    SELECT COUNT(*) INTO v_unpaid_inv
    FROM   invoices
    WHERE  student_id = p_student_id
      AND  school_id  = p_school_id
      AND  status IN ('unpaid', 'partial');

    IF v_unpaid_inv > 0 THEN
        RAISE EXCEPTION
            'Ù„Ø§ ÙŠÙ…ÙƒÙ† Ø£Ø±Ø´ÙØ© Ø§Ù„Ø·Ø§Ù„Ø¨: Ù„Ø¯ÙŠÙ‡ % ÙØ§ØªÙˆØ±Ø©/ÙÙˆØ§ØªÙŠØ± ØºÙŠØ± Ù…Ø³Ø¯Ø¯Ø©.',
            v_unpaid_inv;
    END IF;

    UPDATE students
    SET is_archived = TRUE,
        archived_at = NOW(),
        status      = 'inactive'
    WHERE id        = p_student_id
      AND school_id = p_school_id;

    INSERT INTO audit_logs (school_id, user_id, action, table_name, record_id, new_values)
    VALUES (p_school_id, p_archived_by, 'ARCHIVE', 'students', p_student_id::TEXT,
            jsonb_build_object('is_archived', TRUE, 'archived_at', NOW()));
END;
$$ LANGUAGE plpgsql;

-- Ø£Ø±Ø´ÙØ© Ù…Ø¹Ù„Ù…: Ù„Ø§ Ø­Ù„Ù‚Ø§Øª Ù†Ø´Ø·Ø©
CREATE OR REPLACE FUNCTION archive_teacher(
    p_teacher_id  UUID,
    p_school_id   UUID,
    p_archived_by UUID
) RETURNS VOID AS $$
DECLARE
    v_active_groups INT;
BEGIN
    SELECT COUNT(*) INTO v_active_groups
    FROM   groups
    WHERE  teacher_id  = p_teacher_id
      AND  school_id   = p_school_id
      AND  is_archived = FALSE;

    IF v_active_groups > 0 THEN
        RAISE EXCEPTION
            'Ù„Ø§ ÙŠÙ…ÙƒÙ† Ø£Ø±Ø´ÙØ© Ø§Ù„Ù…Ø¹Ù„Ù…: Ù„Ø¯ÙŠÙ‡ % Ø­Ù„Ù‚Ø©/Ø­Ù„Ù‚Ø§Øª Ù†Ø´Ø·Ø© â€” Ø¹ÙŠÙ‘Ù† Ù…Ø¹Ù„Ù…Ø§Ù‹ Ø¨Ø¯ÙŠÙ„Ø§Ù‹ Ø£ÙˆÙ„Ø§Ù‹.',
            v_active_groups;
    END IF;

    UPDATE teachers
    SET is_archived = TRUE,
        archived_at = NOW(),
        is_active   = FALSE
    WHERE id        = p_teacher_id
      AND school_id = p_school_id;

    INSERT INTO audit_logs (school_id, user_id, action, table_name, record_id, new_values)
    VALUES (p_school_id, p_archived_by, 'ARCHIVE', 'teachers', p_teacher_id::TEXT,
            jsonb_build_object('is_archived', TRUE, 'archived_at', NOW()));
END;
$$ LANGUAGE plpgsql;

-- â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•
--  15. Ù…Ù†Ø¹ DELETE Ø§Ù„Ù…Ø¨Ø§Ø´Ø± Ø¹Ù„Ù‰ Ø§Ù„Ø¬Ø¯Ø§ÙˆÙ„ Ø§Ù„Ø­Ø³Ø§Ø³Ø©
--      (Ø§Ø³ØªØ®Ø¯Ù… archive_student / archive_teacher Ø¨Ø¯Ù„Ø§Ù‹ Ù…Ù†Ù‡)
-- â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•

CREATE OR REPLACE FUNCTION prevent_direct_delete()
RETURNS TRIGGER AS $$
BEGIN
    RAISE EXCEPTION
        'Ø§Ù„Ø­Ø°Ù Ø§Ù„Ù…Ø¨Ø§Ø´Ø± Ù…Ù† "%" Ù…Ø±ÙÙˆØ¶. Ø§Ø³ØªØ®Ø¯Ù… Ø¯ÙˆØ§Ù„ Ø§Ù„Ø£Ø±Ø´ÙØ© Ø§Ù„Ù…Ø®ØµØµØ©. record_id=%',
        TG_TABLE_NAME, OLD.id;
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trg_no_delete_students ON students;
CREATE TRIGGER trg_no_delete_students
    BEFORE DELETE ON students
    FOR EACH ROW EXECUTE FUNCTION prevent_direct_delete();

DROP TRIGGER IF EXISTS trg_no_delete_teachers ON teachers;
CREATE TRIGGER trg_no_delete_teachers
    BEFORE DELETE ON teachers
    FOR EACH ROW EXECUTE FUNCTION prevent_direct_delete();

DROP TRIGGER IF EXISTS trg_no_delete_groups ON groups;
CREATE TRIGGER trg_no_delete_groups
    BEFORE DELETE ON groups
    FOR EACH ROW EXECUTE FUNCTION prevent_direct_delete();

-- â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•
--  Ù…Ù„Ø®Øµ Ù…Ø§ ØªÙ… Ø¥Ù†Ø´Ø§Ø¤Ù‡
-- â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•
--
--  Ø§Ù„Ø¬Ø¯Ø§ÙˆÙ„ (12):
--    teachers Â· groups Â· students Â· student_groups Â· surahs
--    attendance Â· memorization_logs Â· recitation_sessions
--    invoices Â· payments Â· expenses Â· payroll_records Â· audit_logs
--
--  RLS (11 Ø¬Ø¯ÙˆÙ„ ØªØ´ØºÙŠÙ„ÙŠ â€” Ù…Ø§ Ø¹Ø¯Ø§ surahs Ùˆ audit_logs):
--    teachers Â· groups Â· students Â· student_groups
--    attendance Â· memorization_logs Â· recitation_sessions
--    invoices Â· payments Â· expenses Â· payroll_records
--
--  Triggers (set_updated_at):     teachers Â· groups Â· students
--                                 invoices Â· payroll_records
--  Triggers (Ù…Ù†Ø¹ DELETE):         students Â· teachers Â· groups
--  Triggers (Ù…Ù†Ø¹ UPDATE/DELETE):  audit_logs
--
--  Functions: archive_student Â· archive_teacher
--             prevent_direct_delete Â· prevent_audit_mutation
--
--  FKs Ù…Ø±ÙƒÙ‘Ø¨Ø© (id, school_id):
--    student_groups â†’ students Â· groups
--    attendance     â†’ groups
--    memorization_logs   â†’ groups
--    recitation_sessions â†’ groups
--
--  Partial Indexes:
--    uq_sg_one_primary_active  (is_primary=true  AND end_date IS NULL)
--    uq_sg_no_duplicate_active (end_date IS NULL)
--    idx_teachers_active       (is_archived=false)
--    idx_groups_active         (is_archived=false)
--    idx_students_active       (is_archived=false)
--    idx_att_group_date        (group_id IS NOT NULL)
--    idx_mem_group             (group_id IS NOT NULL)
--    idx_rec_group             (group_id IS NOT NULL)

-- â”€â”€ Trigger: auto-fill receipt_number on payments insert â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€
CREATE OR REPLACE FUNCTION fill_receipt_number()
RETURNS TRIGGER LANGUAGE plpgsql AS $$
BEGIN
    IF NEW.receipt_number IS NULL THEN
        NEW.receipt_number := 'RCP-' || TO_CHAR(NOW(), 'YYYYMMDD') || '-' ||
                              UPPER(SUBSTRING(NEW.id::TEXT, 1, 6));
    END IF;
    RETURN NEW;
END;
$$;

DROP TRIGGER IF EXISTS trg_fill_receipt_number ON payments;
CREATE TRIGGER trg_fill_receipt_number
    BEFORE INSERT ON payments
    FOR EACH ROW EXECUTE FUNCTION fill_receipt_number();

-- â”€â”€ Trigger: auto-calculate net_salary on payroll_records â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€
CREATE OR REPLACE FUNCTION calc_net_salary()
RETURNS TRIGGER LANGUAGE plpgsql AS $$
BEGIN
    NEW.net_salary := COALESCE(NEW.base_salary, 0)
                    + COALESCE(NEW.housing_allow, 0)
                    + COALESCE(NEW.transport_allow, 0)
                    + COALESCE(NEW.bonus, 0)
                    - COALESCE(NEW.absence_deduction, 0)
                    - COALESCE(NEW.gosi_deduction, 0)
                    - COALESCE(NEW.other_deduction, 0);
    RETURN NEW;
END;
$$;

DROP TRIGGER IF EXISTS trg_calc_net_salary ON payroll_records;
CREATE TRIGGER trg_calc_net_salary
    BEFORE INSERT OR UPDATE ON payroll_records
    FOR EACH ROW EXECUTE FUNCTION calc_net_salary();


COMMIT;
