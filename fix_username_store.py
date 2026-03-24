with open('internal/store/student.go', 'r', encoding='utf-8') as f:
    content = f.read()

# Fix INSERT
old1 = 'monthly_fee, fee_exemption, notes, gender)\n\t\tVALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)\n\t\tRETURNING id, school_id, full_name, date_of_birth, guardian_name,\n\t\t          enrollment_date, memorized_parts, level_on_entry,\n\t\t          monthly_fee, fee_exemption, status, notes, gender,\n\t\t          is_archived, archived_at, created_at, updated_at\n\t`, schoolID, r.FullName, r.DateOfBirth, r.GuardianName,\n\t\tenrollDate, memorized, r.LevelOnEntry,\n\t\tr.MonthlyFee, exemption, r.Notes, r.Gender,'
new1 = 'monthly_fee, fee_exemption, notes, gender, username)\n\t\tVALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)\n\t\tRETURNING id, school_id, full_name, date_of_birth, guardian_name,\n\t\t          enrollment_date, memorized_parts, level_on_entry,\n\t\t          monthly_fee, fee_exemption, status, notes, gender, username,\n\t\t          is_archived, archived_at, created_at, updated_at\n\t`, schoolID, r.FullName, r.DateOfBirth, r.GuardianName,\n\t\tenrollDate, memorized, r.LevelOnEntry,\n\t\tr.MonthlyFee, exemption, r.Notes, r.Gender, r.Username,'

# Fix CREATE scan
old2 = '&s.MonthlyFee, &s.FeeExemption, &s.Status, &s.Notes, &s.Gender,\n\t\t&s.IsArchived, &s.ArchivedAt, &s.CreatedAt, &s.UpdatedAt,'
new2 = '&s.MonthlyFee, &s.FeeExemption, &s.Status, &s.Notes, &s.Gender, &s.Username,\n\t\t&s.IsArchived, &s.ArchivedAt, &s.CreatedAt, &s.UpdatedAt,'

# Fix GetStudent SELECT
old3 = 'SELECT id, school_id, full_name, date_of_birth, guardian_name,\n\t\t       enrollment_date, memorized_parts, level_on_entry,\n\t\t       monthly_fee, fee_exemption, status, notes, gender,\n\t\t       is_archived, archived_at, created_at, updated_at\n\t\tFROM   students\n\t\tWHERE  id = $1'
new3 = 'SELECT id, school_id, full_name, date_of_birth, guardian_name,\n\t\t       enrollment_date, memorized_parts, level_on_entry,\n\t\t       monthly_fee, fee_exemption, status, notes, gender, username,\n\t\t       is_archived, archived_at, created_at, updated_at\n\t\tFROM   students\n\t\tWHERE  id = $1'

# Fix GetStudent scan
old4 = '&s.MonthlyFee, &s.FeeExemption, &s.Status, &s.Notes, &s.Gender,\n\t\t\t&s.IsArchived, &s.ArchivedAt, &s.CreatedAt, &s.UpdatedAt,'
new4 = '&s.MonthlyFee, &s.FeeExemption, &s.Status, &s.Notes, &s.Gender, &s.Username,\n\t\t\t&s.IsArchived, &s.ArchivedAt, &s.CreatedAt, &s.UpdatedAt,'

# Fix ListStudents SELECT
old5 = 'SELECT s.id, s.school_id, s.full_name, s.date_of_birth, s.guardian_name,\n\t\t       s.enrollment_date, s.memorized_parts, s.level_on_entry,\n\t\t       s.monthly_fee, s.fee_exemption, s.status, s.notes, s.gender,\n\t\t       s.is_archived, s.archived_at, s.created_at, s.updated_at'
new5 = 'SELECT s.id, s.school_id, s.full_name, s.date_of_birth, s.guardian_name,\n\t\t       s.enrollment_date, s.memorized_parts, s.level_on_entry,\n\t\t       s.monthly_fee, s.fee_exemption, s.status, s.notes, s.gender, s.username,\n\t\t       s.is_archived, s.archived_at, s.created_at, s.updated_at'

# Fix ListStudents scan
old6 = '&s.MonthlyFee, &s.FeeExemption, &s.Status, &s.Notes, &s.Gender,\n\t\t\t\t&s.IsArchived, &s.ArchivedAt, &s.CreatedAt, &s.UpdatedAt,'
new6 = '&s.MonthlyFee, &s.FeeExemption, &s.Status, &s.Notes, &s.Gender, &s.Username,\n\t\t\t\t&s.IsArchived, &s.ArchivedAt, &s.CreatedAt, &s.UpdatedAt,'

for i, (old, new) in enumerate([(old1,new1),(old2,new2),(old3,new3),(old4,new4),(old5,new5),(old6,new6)]):
    if old in content:
        content = content.replace(old, new, 1)
        print(f"OK: {i+1}")
    else:
        print(f"NOT FOUND: {i+1}")

with open('internal/store/student.go', 'w', encoding='utf-8') as f:
    f.write(content)
