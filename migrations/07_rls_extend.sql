-- 07_rls_extend.sql — توسيع RLS لباقي الجداول متعددة المدارس
-- يعتمد على نفس منطق 06_rls_final.sql
-- ملاحظة: لا نطبّق RLS على جدول users

DO $$
DECLARE
  t TEXT;
BEGIN
  -- فعّل RLS + FORCE RLS
  FOREACH t IN ARRAY ARRAY['teachers', 'recitation_sessions', 'invoices', 'payments', 'expenses', 'payroll_records', 'audit_logs']
  LOOP
    -- إذا كان الجدول غير موجود (لأي سبب)، تجاهله
    IF to_regclass(t) IS NULL THEN
      RAISE NOTICE 'skip %, table does not exist', t;
      CONTINUE;
    END IF;

    EXECUTE format('ALTER TABLE %I ENABLE ROW LEVEL SECURITY', t);
    EXECUTE format('ALTER TABLE %I FORCE ROW LEVEL SECURITY', t);

    -- سياسة العزل (نفس الاسم المستخدم في 06)
    EXECUTE format('DROP POLICY IF EXISTS tenant_isolation ON %I', t);
    EXECUTE format('DROP POLICY IF EXISTS school_isolation  ON %I', t);

    EXECUTE format($p$
      CREATE POLICY tenant_isolation ON %I
        USING (
          current_setting('app.is_super_admin', true) = '1'
          OR school_id = current_setting('app.school_id', true)::uuid
        )
        WITH CHECK (
          current_setting('app.is_super_admin', true) = '1'
          OR school_id = current_setting('app.school_id', true)::uuid
        )
    $p$, t);
  END LOOP;
END $$;

-- تأكيد أن users بدون RLS
ALTER TABLE users NO FORCE ROW LEVEL SECURITY;
ALTER TABLE users DISABLE ROW LEVEL SECURITY;
