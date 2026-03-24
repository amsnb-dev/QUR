with open('internal/store/student.go', 'r', encoding='utf-8') as f:
    content = f.read()

# Fix CreateStudent to include gender
old1 = 'err := tx.QueryRow(ctx, `\n\t\tINSERT INTO students\n\t\t       (school_id, full_name, date_of_birth, guardian_name,\n\t\t        enrollment_date, memorized_parts, level_on_entry,\n\t\t        monthly_fee, fee_exemption, notes)\n\t\tVALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)\n\t\tRETURNING id, school_id, full_name, date_of_birth, guardian_name,\n\t\t          enrollment_date, memorized_parts, level_on_entry,\n\t\t          monthly_fee, fee_exemption, status, notes,\n\t\t          is_archived, archived_at, created_at, updated_at\n\t`, schoolID, r.FullName, r.DateOfBirth, r.GuardianName,\n\t\tenrollDate, memorized, r.LevelOnEntry,\n\t\tr.MonthlyFee, exemption, r.Notes,'
new1 = 'err := tx.QueryRow(ctx, `\n\t\tINSERT INTO students\n\t\t       (school_id, full_name, date_of_birth, guardian_name,\n\t\t        enrollment_date, memorized_parts, level_on_entry,\n\t\t        monthly_fee, fee_exemption, notes, gender)\n\t\tVALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)\n\t\tRETURNING id, school_id, full_name, date_of_birth, guardian_name,\n\t\t          enrollment_date, memorized_parts, level_on_entry,\n\t\t          monthly_fee, fee_exemption, status, notes, gender,\n\t\t          is_archived, archived_at, created_at, updated_at\n\t`, schoolID, r.FullName, r.DateOfBirth, r.GuardianName,\n\t\tenrollDate, memorized, r.LevelOnEntry,\n\t\tr.MonthlyFee, exemption, r.Notes, r.Gender,'

if old1 in content:
    content = content.replace(old1, new1, 1)
    print("Insert OK")
else:
    print("Insert NOT FOUND")
    idx = content.find("INSERT INTO students")
    print(repr(content[idx:idx+400]))

with open('internal/store/student.go', 'w', encoding='utf-8') as f:
    f.write(content)
