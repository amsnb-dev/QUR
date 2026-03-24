content = open('internal/store/student.go', encoding='utf-8').read()
fixes = [
    # CREATE scan
    ('&s.MonthlyFee, &s.FeeExemption, &s.Status, &s.Notes, &s.Gender, &s.Username,\n\t\t&s.IsArchived',
     '&s.MonthlyFee, &s.FeeExemption, &s.Status, &s.Notes, &s.Gender, &s.Username, &s.PasswordPlain,\n\t\t&s.IsArchived'),
    # GetStudent SELECT
    ('monthly_fee, fee_exemption, status, notes, gender, username,\n\t\t       is_archived, archived_at, created_at, updated_at\n\t\tFROM   students\n\t\tWHERE  id = $1',
     'monthly_fee, fee_exemption, status, notes, gender, username, password_plain,\n\t\t       is_archived, archived_at, created_at, updated_at\n\t\tFROM   students\n\t\tWHERE  id = $1'),
    # GetStudent scan
    ('&s.MonthlyFee, &s.FeeExemption, &s.Status, &s.Notes, &s.Gender, &s.Username,\n\t\t\t&s.IsArchived',
     '&s.MonthlyFee, &s.FeeExemption, &s.Status, &s.Notes, &s.Gender, &s.Username, &s.PasswordPlain,\n\t\t\t&s.IsArchived'),
    # ListStudents SELECT
    ('s.monthly_fee, s.fee_exemption, s.status, s.notes, s.gender, s.username,\n\t\t       s.is_archived',
     's.monthly_fee, s.fee_exemption, s.status, s.notes, s.gender, s.username, s.password_plain,\n\t\t       s.is_archived'),
    # ListStudents scan
    ('&s.MonthlyFee, &s.FeeExemption, &s.Status, &s.Notes, &s.Gender, &s.Username,\n\t\t\t\t&s.IsArchived',
     '&s.MonthlyFee, &s.FeeExemption, &s.Status, &s.Notes, &s.Gender, &s.Username, &s.PasswordPlain,\n\t\t\t\t&s.IsArchived'),
    # UpdateStudent RETURNING
    ('monthly_fee, fee_exemption, status, notes, gender, username,\n\t\t          is_archived, archived_at, created_at, updated_at',
     'monthly_fee, fee_exemption, status, notes, gender, username, password_plain,\n\t\t          is_archived, archived_at, created_at, updated_at'),
    # UpdateStudent scan
    ('&s.MonthlyFee, &s.FeeExemption, &s.Status, &s.Notes, &s.Gender, &s.Username,\n\t\t&s.IsArchived',
     '&s.MonthlyFee, &s.FeeExemption, &s.Status, &s.Notes, &s.Gender, &s.Username, &s.PasswordPlain,\n\t\t&s.IsArchived'),
]
for i, (old, new) in enumerate(fixes):
    if old in content:
        content = content.replace(old, new, 1)
        print(f"OK: {i+1}")
    else:
        print(f"NOT FOUND: {i+1}")
open('internal/store/student.go', 'w', encoding='utf-8').write(content)
