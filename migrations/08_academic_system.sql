-- ============================================================
--  08_academic_system.sql
--  النظام الأكاديمي: السنوات الدراسية، التسجيل، الاختبارات
--  Multi-Tenant · RLS · Idempotent
-- ============================================================
--  المتطلبات: 00_init + 01_schema + 03_core_mvp
--  (schools · users · students · groups · teachers موجودة)
--  لا يعدّل أي جدول موجود — إضافة فقط
-- ============================================================

BEGIN;

-- ════════════════════════════════════════════════════════════
--  1. academic_years
--     السنة الدراسية لكل مدرسة (per-school)
--     مثال: "2024-2025"، من 2024-09-01 إلى 2025-06-30
-- ════════════════════════════════════════════════════════════

CREATE TABLE IF NOT EXISTS academic_years (
    id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    school_id   UUID        NOT NULL REFERENCES schools (id) ON DELETE RESTRICT,
    name        TEXT        NOT NULL,               -- "2024-2025"
    start_date  DATE        NOT NULL,
    end_date    DATE        NOT NULL,
    is_current  BOOLEAN     NOT NULL DEFAULT FALSE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_academic_year_school_name UNIQUE (school_id, name),
    CONSTRAINT uq_academic_year_id_school   UNIQUE (id, school_id),
    CONSTRAINT chk_academic_year_dates      CHECK (end_date > start_date)
);

-- سنة واحدة فقط is_current = TRUE لكل مدرسة
CREATE UNIQUE INDEX IF NOT EXISTS uq_academic_year_current
    ON academic_years (school_id)
    WHERE is_current = TRUE;

CREATE OR REPLACE TRIGGER trg_academic_years_updated_at
    BEFORE UPDATE ON academic_years
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE INDEX IF NOT EXISTS idx_academic_years_school
    ON academic_years (school_id, start_date DESC);

-- RLS
ALTER TABLE academic_years ENABLE ROW LEVEL SECURITY;

DROP POLICY IF EXISTS school_isolation ON academic_years;
CREATE POLICY school_isolation ON academic_years
    USING (
        school_id = NULLIF(current_setting('app.school_id', TRUE), '')::UUID
        OR COALESCE(current_setting('app.school_id', TRUE), '') = ''
    );

-- ════════════════════════════════════════════════════════════
--  2. student_enrollments
--     تسجيل الطالب في سنة دراسية
--     طالب واحد → سنة واحدة → تسجيل واحد
-- ════════════════════════════════════════════════════════════

CREATE TABLE IF NOT EXISTS student_enrollments (
    id               UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    school_id        UUID        NOT NULL REFERENCES schools (id) ON DELETE RESTRICT,
    student_id       UUID        NOT NULL,
    academic_year_id UUID        NOT NULL,
    -- مستوى الطالب عند التسجيل في هذه السنة
    level_at_entry   TEXT        CHECK (level_at_entry IN ('beginner', 'intermediate', 'advanced')),
    -- عدد الأجزاء المحفوظة عند بداية السنة
    parts_at_entry   NUMERIC(4,1) NOT NULL DEFAULT 0
                     CHECK (parts_at_entry BETWEEN 0 AND 30),
    -- الهدف للسنة
    target_parts     NUMERIC(4,1) CHECK (target_parts BETWEEN 0 AND 30),
    status           TEXT        NOT NULL DEFAULT 'active'
                     CHECK (status IN ('active', 'completed', 'withdrawn')),
    notes            TEXT,
    enrolled_by      UUID        NOT NULL REFERENCES users (id),
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    -- FK مركب: طالب + مدرسة من نفس المدرسة
    CONSTRAINT fk_enrollment_student_school
        FOREIGN KEY (student_id, school_id)
        REFERENCES students (id, school_id) ON DELETE RESTRICT,
    -- FK مركب: سنة + مدرسة من نفس المدرسة
    CONSTRAINT fk_enrollment_year_school
        FOREIGN KEY (academic_year_id, school_id)
        REFERENCES academic_years (id, school_id) ON DELETE RESTRICT,
    -- طالب مرة واحدة في نفس السنة
    CONSTRAINT uq_enrollment_student_year
        UNIQUE (school_id, student_id, academic_year_id)
);

CREATE OR REPLACE TRIGGER trg_student_enrollments_updated_at
    BEFORE UPDATE ON student_enrollments
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE INDEX IF NOT EXISTS idx_enrollment_school_year
    ON student_enrollments (school_id, academic_year_id);
CREATE INDEX IF NOT EXISTS idx_enrollment_student
    ON student_enrollments (student_id);
CREATE INDEX IF NOT EXISTS idx_enrollment_active
    ON student_enrollments (school_id, academic_year_id)
    WHERE status = 'active';

-- RLS
ALTER TABLE student_enrollments ENABLE ROW LEVEL SECURITY;

DROP POLICY IF EXISTS school_isolation ON student_enrollments;
CREATE POLICY school_isolation ON student_enrollments
    USING (
        school_id = NULLIF(current_setting('app.school_id', TRUE), '')::UUID
        OR COALESCE(current_setting('app.school_id', TRUE), '') = ''
    );

-- ════════════════════════════════════════════════════════════
--  3. exams
--     الاختبارات الدورية لكل مدرسة
-- ════════════════════════════════════════════════════════════

CREATE TABLE IF NOT EXISTS exams (
    id               UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    school_id        UUID        NOT NULL REFERENCES schools (id) ON DELETE RESTRICT,
    academic_year_id UUID        NOT NULL,
    group_id         UUID,                          -- اختبار لحلقة معينة أو لكل المدرسة
    name             TEXT        NOT NULL,          -- "اختبار منتصف الفصل الأول"
    exam_type        TEXT        NOT NULL DEFAULT 'periodic'
                     CHECK (exam_type IN ('periodic', 'final', 'placement', 'competition')),
    exam_date        DATE        NOT NULL,
    -- نطاق الاختبار
    from_surah       SMALLINT    REFERENCES surahs (number),
    to_surah         SMALLINT    REFERENCES surahs (number),
    max_score        NUMERIC(6,2) NOT NULL DEFAULT 100,
    pass_score       NUMERIC(6,2) NOT NULL DEFAULT 60,
    notes            TEXT,
    created_by       UUID        NOT NULL REFERENCES users (id),
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_exam_id_school UNIQUE (id, school_id),
    CONSTRAINT chk_exam_scores   CHECK (pass_score <= max_score AND max_score > 0),
    -- FK مركب: سنة + مدرسة
    CONSTRAINT fk_exam_year_school
        FOREIGN KEY (academic_year_id, school_id)
        REFERENCES academic_years (id, school_id) ON DELETE RESTRICT,
    -- FK مركب: حلقة + مدرسة (nullable)
    CONSTRAINT fk_exam_group_school
        FOREIGN KEY (group_id, school_id)
        REFERENCES groups (id, school_id) ON DELETE SET NULL
);

CREATE OR REPLACE TRIGGER trg_exams_updated_at
    BEFORE UPDATE ON exams
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE INDEX IF NOT EXISTS idx_exams_school_year
    ON exams (school_id, academic_year_id, exam_date DESC);
CREATE INDEX IF NOT EXISTS idx_exams_group
    ON exams (group_id, exam_date DESC)
    WHERE group_id IS NOT NULL;

-- RLS
ALTER TABLE exams ENABLE ROW LEVEL SECURITY;

DROP POLICY IF EXISTS school_isolation ON exams;
CREATE POLICY school_isolation ON exams
    USING (
        school_id = NULLIF(current_setting('app.school_id', TRUE), '')::UUID
        OR COALESCE(current_setting('app.school_id', TRUE), '') = ''
    );

-- ════════════════════════════════════════════════════════════
--  4. exam_results
--     نتائج الطلاب في الاختبارات
-- ════════════════════════════════════════════════════════════

CREATE TABLE IF NOT EXISTS exam_results (
    id          UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    school_id   UUID         NOT NULL REFERENCES schools (id) ON DELETE RESTRICT,
    exam_id     UUID         NOT NULL,
    student_id  UUID         NOT NULL,
    score       NUMERIC(6,2) NOT NULL CHECK (score >= 0),
    -- تفاصيل الدرجات
    hafz_score    NUMERIC(6,2) CHECK (hafz_score    >= 0),
    tajweed_score NUMERIC(6,2) CHECK (tajweed_score >= 0),
    -- مجتاز / راسب يُحسب تلقائياً في التطبيق
    is_absent   BOOLEAN      NOT NULL DEFAULT FALSE,
    notes       TEXT,
    recorded_by UUID         NOT NULL REFERENCES users (id),
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    -- FK مركب: امتحان + مدرسة
    CONSTRAINT fk_result_exam_school
        FOREIGN KEY (exam_id, school_id)
        REFERENCES exams (id, school_id) ON DELETE RESTRICT,
    -- FK مركب: طالب + مدرسة
    CONSTRAINT fk_result_student_school
        FOREIGN KEY (student_id, school_id)
        REFERENCES students (id, school_id) ON DELETE RESTRICT,
    -- نتيجة واحدة لكل طالب في كل اختبار
    CONSTRAINT uq_exam_result_student
        UNIQUE (school_id, exam_id, student_id)
);

CREATE OR REPLACE TRIGGER trg_exam_results_updated_at
    BEFORE UPDATE ON exam_results
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE INDEX IF NOT EXISTS idx_exam_results_exam
    ON exam_results (exam_id, score DESC);
CREATE INDEX IF NOT EXISTS idx_exam_results_student
    ON exam_results (student_id, created_at DESC);

-- RLS
ALTER TABLE exam_results ENABLE ROW LEVEL SECURITY;

DROP POLICY IF EXISTS school_isolation ON exam_results;
CREATE POLICY school_isolation ON exam_results
    USING (
        school_id = NULLIF(current_setting('app.school_id', TRUE), '')::UUID
        OR COALESCE(current_setting('app.school_id', TRUE), '') = ''
    );

-- ════════════════════════════════════════════════════════════
--  5. holidays
--     الإجازات والعطل الرسمية لكل مدرسة
-- ════════════════════════════════════════════════════════════

CREATE TABLE IF NOT EXISTS holidays (
    id               UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    school_id        UUID        NOT NULL REFERENCES schools (id) ON DELETE RESTRICT,
    academic_year_id UUID,                          -- NULL = عطلة دائمة
    name             TEXT        NOT NULL,
    start_date       DATE        NOT NULL,
    end_date         DATE        NOT NULL,
    holiday_type     TEXT        NOT NULL DEFAULT 'official'
                     CHECK (holiday_type IN ('official', 'school', 'emergency')),
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_holiday_dates CHECK (end_date >= start_date),
    -- FK مركب: سنة + مدرسة (nullable)
    CONSTRAINT fk_holiday_year_school
        FOREIGN KEY (academic_year_id, school_id)
        REFERENCES academic_years (id, school_id) ON DELETE SET NULL
);

CREATE INDEX IF NOT EXISTS idx_holidays_school
    ON holidays (school_id, start_date);

-- RLS
ALTER TABLE holidays ENABLE ROW LEVEL SECURITY;

DROP POLICY IF EXISTS school_isolation ON holidays;
CREATE POLICY school_isolation ON holidays
    USING (
        school_id = NULLIF(current_setting('app.school_id', TRUE), '')::UUID
        OR COALESCE(current_setting('app.school_id', TRUE), '') = ''
    );

-- ════════════════════════════════════════════════════════════
--  ملخص ما تم إنشاؤه
-- ════════════════════════════════════════════════════════════
--
--  الجداول (5):
--    academic_years · student_enrollments · exams
--    exam_results · holidays
--
--  RLS (5 جداول — جميعها)
--
--  Triggers (set_updated_at):
--    academic_years · student_enrollments · exams · exam_results
--
--  FKs مركبة (id, school_id):
--    student_enrollments → students · academic_years
--    exams               → academic_years · groups
--    exam_results        → exams · students
--    holidays            → academic_years
--
--  Partial Indexes:
--    uq_academic_year_current   (is_current = TRUE)
--    idx_enrollment_active      (status = 'active')
--    idx_exams_group            (group_id IS NOT NULL)
--
--  لا تعديل على أي جدول موجود
--  لا تعارض مع: payroll_records (موجود في 03) · memorization_logs · attendance
--
-- ════════════════════════════════════════════════════════════

COMMIT;
