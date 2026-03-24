with open('internal/store/student.go', 'r', encoding='utf-8') as f:
    content = f.read()

# Fix Scan in CreateStudent
old1 = '&s.MonthlyFee, &s.FeeExemption, &s.Status, &s.Notes,\n\t\t&s.IsArchived, &s.ArchivedAt, &s.CreatedAt, &s.UpdatedAt,'
new1 = '&s.MonthlyFee, &s.FeeExemption, &s.Status, &s.Notes, &s.Gender,\n\t\t&s.IsArchived, &s.ArchivedAt, &s.CreatedAt, &s.UpdatedAt,'

# Fix SELECT in GetStudent and UpdateStudent RETURNING
old2 = 'SELECT id, school_id, full_name, date_of_birth, guardian_name,\n\t\t       enrollment_date, memorized_parts, level_on_entry,\n\t\t       monthly_fee, fee_exemption, status, notes,\n\t\t       is_archived, archived_at, created_at, updated_at\n\t\tFROM   students\n\t\tWHERE  id = $1'
new2 = 'SELECT id, school_id, full_name, date_of_birth, guardian_name,\n\t\t       enrollment_date, memorized_parts, level_on_entry,\n\t\t       monthly_fee, fee_exemption, status, notes, gender,\n\t\t       is_archived, archived_at, created_at, updated_at\n\t\tFROM   students\n\t\tWHERE  id = $1'

count = 0
for old, new in [(old1, new1), (old2, new2)]:
    if old in content:
        content = content.replace(old, new)
        count += 1
        print(f"OK: {count}")
    else:
        print(f"NOT FOUND: {repr(old[:50])}")

with open('internal/store/student.go', 'w', encoding='utf-8') as f:
    f.write(content)
