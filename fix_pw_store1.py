content = open('internal/store/student.go', encoding='utf-8').read()
old1 = 'monthly_fee, fee_exemption, notes, gender, username)\n\t\tVALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)\n\t\tRETURNING id, school_id, full_name, date_of_birth, guardian_name,\n\t\t          enrollment_date, memorized_parts, level_on_entry,\n\t\t          monthly_fee, fee_exemption, status, notes, gender, username,\n\t\t          is_archived, archived_at, created_at, updated_at\n\t`, schoolID, r.FullName, r.DateOfBirth, r.GuardianName,\n\t\tenrollDate, memorized, r.LevelOnEntry,\n\t\tr.MonthlyFee, exemption, r.Notes, r.Gender, r.Username,'
new1 = 'monthly_fee, fee_exemption, notes, gender, username, password_plain)\n\t\tVALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)\n\t\tRETURNING id, school_id, full_name, date_of_birth, guardian_name,\n\t\t          enrollment_date, memorized_parts, level_on_entry,\n\t\t          monthly_fee, fee_exemption, status, notes, gender, username, password_plain,\n\t\t          is_archived, archived_at, created_at, updated_at\n\t`, schoolID, r.FullName, r.DateOfBirth, r.GuardianName,\n\t\tenrollDate, memorized, r.LevelOnEntry,\n\t\tr.MonthlyFee, exemption, r.Notes, r.Gender, r.Username, r.PasswordPlain,'
if old1 in content:
    content = content.replace(old1, new1, 1)
    print("OK")
else:
    print("NOT FOUND")
    idx = content.find('monthly_fee, fee_exemption, notes, gender')
    print(repr(content[idx:idx+200]))
open('internal/store/student.go', 'w', encoding='utf-8').write(content)
