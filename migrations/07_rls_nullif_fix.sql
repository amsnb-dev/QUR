-- Fix RLS policies to use NULLIF for empty school_id (super_admin support)
DROP POLICY IF EXISTS tenant_isolation ON students;
CREATE POLICY tenant_isolation ON students
  USING (current_setting('app.is_super_admin',true)='1' OR school_id=NULLIF(current_setting('app.school_id',true),'')::uuid)
  WITH CHECK (current_setting('app.is_super_admin',true)='1' OR school_id=NULLIF(current_setting('app.school_id',true),'')::uuid);

DROP POLICY IF EXISTS tenant_isolation ON teachers;
CREATE POLICY tenant_isolation ON teachers
  USING (current_setting('app.is_super_admin',true)='1' OR school_id=NULLIF(current_setting('app.school_id',true),'')::uuid)
  WITH CHECK (current_setting('app.is_super_admin',true)='1' OR school_id=NULLIF(current_setting('app.school_id',true),'')::uuid);

DROP POLICY IF EXISTS tenant_isolation ON groups;
CREATE POLICY tenant_isolation ON groups
  USING (current_setting('app.is_super_admin',true)='1' OR school_id=NULLIF(current_setting('app.school_id',true),'')::uuid)
  WITH CHECK (current_setting('app.is_super_admin',true)='1' OR school_id=NULLIF(current_setting('app.school_id',true),'')::uuid);
