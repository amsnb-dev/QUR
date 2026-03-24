DO $$
DECLARE
  v_school_a  UUID := 'a0000000-0000-0000-0000-000000000001';
  v_school_b  UUID := 'b0000000-0000-0000-0000-000000000002';
  v_super_id  UUID := 'c0000000-0000-0000-0000-000000000001';
  v_admin_a   UUID := 'c0000000-0000-0000-0000-000000000002';
  v_admin_b   UUID := 'c0000000-0000-0000-0000-000000000003';
  v_hash      TEXT := '$2a$12$LQv3c1yqBWVHxkd0LHAkCOYz6TtxMQJqhN8/LewdBPj2NdNMKYk2';
BEGIN
  INSERT INTO schools (id, name, city)
  VALUES (v_school_a, 'مدرسة نور القرآن', 'الجزائر العاصمة')
  ON CONFLICT (id) DO NOTHING;

  INSERT INTO schools (id, name, city)
  VALUES (v_school_b, 'مدرسة الاختبار B', 'وهران')
  ON CONFLICT (id) DO NOTHING;

  INSERT INTO users (id, school_id, role_id, full_name, email, password_hash)
  VALUES (
    v_super_id, NULL,
    (SELECT id FROM roles WHERE name = 'super_admin'),
    'Super Admin', 'super@quran.dev', v_hash
  ) ON CONFLICT DO NOTHING;

  INSERT INTO users (id, school_id, role_id, full_name, email, password_hash)
  VALUES (
    v_admin_a, v_school_a,
    (SELECT id FROM roles WHERE name = 'school_admin'),
    'أحمد المدير', 'admin@quran.dev', v_hash
  ) ON CONFLICT DO NOTHING;

  INSERT INTO users (id, school_id, role_id, full_name, email, password_hash)
  VALUES (
    v_admin_b, v_school_b,
    (SELECT id FROM roles WHERE name = 'school_admin'),
    'محمد المدير', 'adminb@quran.dev', v_hash
  ) ON CONFLICT DO NOTHING;
END $$;