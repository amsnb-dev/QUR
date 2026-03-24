with open('internal/store/student.go', 'r', encoding='utf-8') as f:
    content = f.read()

# Fix ListStudents SELECT
old1 = 'SELECT s.id, s.school_id, s.full_name, s.date_of_birth, s.guardian_name,\n\t\t       s.enrollment_date, s.memorized_parts, s.level_on_entry,\n\t\t       s.monthly_fee, s.fee_exemption, s.status, s.notes,\n\t\t       s.is_archived, s.archived_at, s.created_at, s.updated_at'
new1 = 'SELECT s.id, s.school_id, s.full_name, s.date_of_birth, s.guardian_name,\n\t\t       s.enrollment_date, s.memorized_parts, s.level_on_entry,\n\t\t       s.monthly_fee, s.fee_exemption, s.status, s.notes, s.gender,\n\t\t       s.is_archived, s.archived_at, s.created_at, s.updated_at'

# Fix UpdateStudent SET and RETURNING
old2 = 'SET    full_name       = $2,\n\t\t       date_of_birth   = $3,\n\t\t       guardian_name   = $4,\n\t\t       memorized_parts = $5,\n\t\t       level_on_entry  = $6,\n\t\t       monthly_fee     = $7,\n\t\t       fee_exemption   = $8,\n\t\t       status          = $9,\n\t\t       notes           = $10\n\t\tWHERE  id = $1\n\t\tRETURNING id, school_id, full_name, date_of_birth, guardian_name,\n\t\t          enrollment_date, memorized_parts, level_on_entry,\n\t\t          monthly_fee, fee_exemption, status, notes,\n\t\t          is_archived, archived_at, created_at, updated_at\n\t`, id, fullName, dob, guardian, memorized, levelEntry, fee, exemption, status, notes,'
new2 = 'SET    full_name       = $2,\n\t\t       date_of_birth   = $3,\n\t\t       guardian_name   = $4,\n\t\t       memorized_parts = $5,\n\t\t       level_on_entry  = $6,\n\t\t       monthly_fee     = $7,\n\t\t       fee_exemption   = $8,\n\t\t       status          = $9,\n\t\t       notes           = $10,\n\t\t       gender          = $11\n\t\tWHERE  id = $1\n\t\tRETURNING id, school_id, full_name, date_of_birth, guardian_name,\n\t\t          enrollment_date, memorized_parts, level_on_entry,\n\t\t          monthly_fee, fee_exemption, status, notes, gender,\n\t\t          is_archived, archived_at, created_at, updated_at\n\t`, id, fullName, dob, guardian, memorized, levelEntry, fee, exemption, status, notes, gender,'

for i, (old, new) in enumerate([(old1, new1), (old2, new2)]):
    if old in content:
        content = content.replace(old, new)
        print(f"OK: {i+1}")
    else:
        print(f"NOT FOUND: {i+1} - {repr(old[:60])}")

with open('internal/store/student.go', 'w', encoding='utf-8') as f:
    f.write(content)
