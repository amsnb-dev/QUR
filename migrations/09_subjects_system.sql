-- ============================================================
-- Migration 09 FIX: Drop incomplete tables and recreate clean
-- ============================================================
BEGIN;

DROP TABLE IF EXISTS subject_sessions  CASCADE;
DROP TABLE IF EXISTS student_subjects  CASCADE;
DROP TABLE IF EXISTS subject_levels    CASCADE;
DROP TABLE IF EXISTS subjects          CASCADE;

-- 1. subjects
CREATE TABLE subjects (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    school_id   UUID NOT NULL REFERENCES schools(id) ON DELETE CASCADE,
    name        TEXT NOT NULL,
    description TEXT,
    category    TEXT NOT NULL DEFAULT 'quran'
                CHECK (category IN ('quran','arabic','islamic','other')),
    color       TEXT DEFAULT '#1a7a4e',
    icon        TEXT DEFAULT '📖',
    is_active   BOOLEAN NOT NULL DEFAULT TRUE,
    sort_order  INTEGER NOT NULL DEFAULT 0,
    created_by  UUID REFERENCES users(id) ON DELETE SET NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (school_id, name),
    UNIQUE (id, school_id)
);
CREATE TRIGGER trg_subjects_updated_at BEFORE UPDATE ON subjects FOR EACH ROW EXECUTE FUNCTION set_updated_at();
ALTER TABLE subjects ENABLE ROW LEVEL SECURITY;
ALTER TABLE subjects FORCE ROW LEVEL SECURITY;
CREATE POLICY subjects_tenant ON subjects USING (school_id = NULLIF(current_setting('app.school_id', TRUE), '')::UUID OR COALESCE(current_setting('app.school_id', TRUE), '') = '');

-- 2. subject_levels
CREATE TABLE subject_levels (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    school_id   UUID NOT NULL REFERENCES schools(id) ON DELETE CASCADE,
    subject_id  UUID NOT NULL,
    name        TEXT NOT NULL,
    description TEXT,
    order_index INTEGER NOT NULL DEFAULT 0,
    criteria    TEXT,
    is_active   BOOLEAN NOT NULL DEFAULT TRUE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    FOREIGN KEY (subject_id, school_id) REFERENCES subjects(id, school_id) ON DELETE CASCADE,
    UNIQUE (school_id, subject_id, name),
    UNIQUE (id, school_id)
);
CREATE TRIGGER trg_subject_levels_updated_at BEFORE UPDATE ON subject_levels FOR EACH ROW EXECUTE FUNCTION set_updated_at();
ALTER TABLE subject_levels ENABLE ROW LEVEL SECURITY;
ALTER TABLE subject_levels FORCE ROW LEVEL SECURITY;
CREATE POLICY subject_levels_tenant ON subject_levels USING (school_id = NULLIF(current_setting('app.school_id', TRUE), '')::UUID OR COALESCE(current_setting('app.school_id', TRUE), '') = '');

-- 3. student_subjects
CREATE TABLE student_subjects (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    school_id        UUID NOT NULL REFERENCES schools(id) ON DELETE CASCADE,
    student_id       UUID NOT NULL,
    subject_id       UUID NOT NULL,
    current_level_id UUID,
    status      TEXT NOT NULL DEFAULT 'active' CHECK (status IN ('active','paused','completed','dropped')),
    started_at  DATE NOT NULL DEFAULT CURRENT_DATE,
    notes       TEXT,
    assigned_by UUID REFERENCES users(id) ON DELETE SET NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    FOREIGN KEY (student_id, school_id) REFERENCES students(id, school_id),
    FOREIGN KEY (subject_id, school_id) REFERENCES subjects(id, school_id),
    UNIQUE (school_id, student_id, subject_id)
);
CREATE TRIGGER trg_student_subjects_updated_at BEFORE UPDATE ON student_subjects FOR EACH ROW EXECUTE FUNCTION set_updated_at();
ALTER TABLE student_subjects ENABLE ROW LEVEL SECURITY;
ALTER TABLE student_subjects FORCE ROW LEVEL SECURITY;
CREATE POLICY student_subjects_tenant ON student_subjects USING (school_id = NULLIF(current_setting('app.school_id', TRUE), '')::UUID OR COALESCE(current_setting('app.school_id', TRUE), '') = '');
CREATE INDEX idx_student_subjects_student ON student_subjects(school_id, student_id);
CREATE INDEX idx_student_subjects_subject ON student_subjects(school_id, subject_id);

-- 4. subject_sessions
CREATE TABLE subject_sessions (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    school_id        UUID NOT NULL REFERENCES schools(id) ON DELETE CASCADE,
    student_id       UUID NOT NULL,
    subject_id       UUID NOT NULL,
    teacher_id       UUID,
    session_date     DATE NOT NULL DEFAULT CURRENT_DATE,
    content          TEXT,
    pages_count      NUMERIC(5,1),
    duration_minutes INTEGER,
    performance      TEXT CHECK (performance IN ('excellent','good','average','weak','absent')),
    score            NUMERIC(5,2),
    level_id         UUID,
    notes            TEXT,
    recorded_by      UUID REFERENCES users(id) ON DELETE SET NULL,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    FOREIGN KEY (student_id, school_id) REFERENCES students(id, school_id),
    FOREIGN KEY (subject_id, school_id) REFERENCES subjects(id, school_id)
);
CREATE TRIGGER trg_subject_sessions_updated_at BEFORE UPDATE ON subject_sessions FOR EACH ROW EXECUTE FUNCTION set_updated_at();
ALTER TABLE subject_sessions ENABLE ROW LEVEL SECURITY;
ALTER TABLE subject_sessions FORCE ROW LEVEL SECURITY;
CREATE POLICY subject_sessions_tenant ON subject_sessions USING (school_id = NULLIF(current_setting('app.school_id', TRUE), '')::UUID OR COALESCE(current_setting('app.school_id', TRUE), '') = '');
CREATE INDEX idx_subject_sessions_student ON subject_sessions(school_id, student_id, session_date DESC);
CREATE INDEX idx_subject_sessions_subject ON subject_sessions(school_id, subject_id, session_date DESC);
CREATE INDEX idx_subject_sessions_teacher ON subject_sessions(school_id, teacher_id, session_date DESC);
CREATE INDEX idx_subject_sessions_date    ON subject_sessions(school_id, session_date DESC);

-- 5. Link exams to subjects
ALTER TABLE exams ADD COLUMN IF NOT EXISTS subject_id UUID REFERENCES subjects(id) ON DELETE SET NULL;
CREATE INDEX IF NOT EXISTS idx_exams_subject ON exams(school_id, subject_id) WHERE subject_id IS NOT NULL;

-- 6. Grants
GRANT SELECT, INSERT, UPDATE, DELETE ON subjects         TO quran_app;
GRANT SELECT, INSERT, UPDATE, DELETE ON subject_levels   TO quran_app;
GRANT SELECT, INSERT, UPDATE, DELETE ON student_subjects TO quran_app;
GRANT SELECT, INSERT, UPDATE, DELETE ON subject_sessions TO quran_app;

-- 7. Seed School A
INSERT INTO subjects (school_id, name, category, icon, color, sort_order, created_by) VALUES
    ('a0000000-0000-0000-0000-000000000001', 'حفظ القرآن الكريم', 'quran',   '📖', '#1a7a4e', 1, (SELECT id FROM users WHERE email='admin@quran.dev' LIMIT 1)),
    ('a0000000-0000-0000-0000-000000000001', 'التجويد',            'quran',   '🎙️', '#1a5276', 2, (SELECT id FROM users WHERE email='admin@quran.dev' LIMIT 1)),
    ('a0000000-0000-0000-0000-000000000001', 'أحكام التجويد',      'quran',   '📚', '#6c3483', 3, (SELECT id FROM users WHERE email='admin@quran.dev' LIMIT 1)),
    ('a0000000-0000-0000-0000-000000000001', 'التفسير',             'quran',   '🔍', '#784212', 4, (SELECT id FROM users WHERE email='admin@quran.dev' LIMIT 1)),
    ('a0000000-0000-0000-0000-000000000001', 'اللغة العربية',       'arabic',  '✍️', '#b7950b', 5, (SELECT id FROM users WHERE email='admin@quran.dev' LIMIT 1)),
    ('a0000000-0000-0000-0000-000000000001', 'الفقه',               'islamic', '⚖️', '#922b21', 6, (SELECT id FROM users WHERE email='admin@quran.dev' LIMIT 1))
ON CONFLICT (school_id, name) DO NOTHING;

WITH s AS (SELECT id FROM subjects WHERE school_id='a0000000-0000-0000-0000-000000000001' AND name='حفظ القرآن الكريم')
INSERT INTO subject_levels (school_id, subject_id, name, order_index, criteria)
SELECT 'a0000000-0000-0000-0000-000000000001', s.id, lvl.name, lvl.ord, lvl.crit
FROM s, (VALUES ('مبتدئ',1,'أقل من 3 أجزاء'),('جزء عمّ',2,'حفظ جزء عمّ كاملاً'),('ربع القرآن',3,'7.5 جزء'),('نصف القرآن',4,'15 جزء'),('ثلاثة أرباع',5,'22.5 جزء'),('حافظ القرآن',6,'30 جزءاً')) AS lvl(name,ord,crit)
ON CONFLICT DO NOTHING;

WITH s AS (SELECT id FROM subjects WHERE school_id='a0000000-0000-0000-0000-000000000001' AND name='التجويد')
INSERT INTO subject_levels (school_id, subject_id, name, order_index)
SELECT 'a0000000-0000-0000-0000-000000000001', s.id, lvl.name, lvl.ord
FROM s, (VALUES ('مبتدئ',1),('متوسط',2),('متقدم',3),('ممتاز',4)) AS lvl(name,ord)
ON CONFLICT DO NOTHING;

COMMIT;
