import subprocess, json

school_id = 'a0000000-0000-0000-0000-000000000001'
roles = [
    ('مدير',  'admin',      True,  {"students":"write","teachers":"write","groups":"write","exams":"write","attendance":"write","fees":"write","payroll":"write","guardians":"write","settings":"write","staff":"write","reports":"write"}),
    ('مشرف',  'supervisor', True,  {"students":"write","teachers":"write","groups":"write","exams":"write","attendance":"write","fees":"read","payroll":"none","guardians":"write","settings":"none","staff":"none","reports":"read"}),
    ('محاسب', 'accountant', True,  {"students":"read","teachers":"read","groups":"read","exams":"none","attendance":"read","fees":"write","payroll":"write","guardians":"read","settings":"none","staff":"none","reports":"write"}),
    ('موظف',  'staff',      True,  {"students":"read","teachers":"read","groups":"read","exams":"read","attendance":"read","fees":"none","payroll":"none","guardians":"read","settings":"none","staff":"none","reports":"none"}),
]

for name, slug, is_sys, perms in roles:
    p = json.dumps(perms, ensure_ascii=False)
    sql = f"INSERT INTO school_roles (school_id,name,slug,is_system,permissions) VALUES ('{school_id}','{name}','{slug}',{str(is_sys).lower()},'{p}') ON CONFLICT (school_id,slug) DO NOTHING;"
    r = subprocess.run(['docker','exec','-i','quran_db','psql','-U','quran','-d','quran_school','-c',sql], capture_output=True, text=True)
    print(name, r.stdout.strip(), r.stderr.strip())

r = subprocess.run(['docker','exec','-i','quran_db','psql','-U','quran','-d','quran_school','-c',"SELECT name,slug FROM school_roles WHERE school_id='a0000000-0000-0000-0000-000000000001';"], capture_output=True, text=True)
print(r.stdout)
