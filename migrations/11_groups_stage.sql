-- ============================================================
-- Migration 11: Groups — add stage, drop capacity
-- ============================================================
-- Run: Get-Content migrations\11_groups_stage.sql | docker exec -i quran_db psql -U quran -d quran_school

BEGIN;

-- إضافة حقل المرحلة
ALTER TABLE groups ADD COLUMN IF NOT EXISTS stage TEXT
    CHECK (stage IN ('primary','middle','secondary','adult','mixed','custom'));

-- حذف طاقة الاستيعاب
ALTER TABLE groups DROP COLUMN IF EXISTS capacity;

-- جعل teacher_id اختياري (إزالة NOT NULL إن وجدت)
ALTER TABLE groups ALTER COLUMN teacher_id DROP NOT NULL;

COMMIT;
