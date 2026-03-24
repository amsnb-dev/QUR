content = open('internal/store/student.go', encoding='utf-8').read()
old = 'monthly_fee, fee_exemption, status, notes, gender, username, password_plain,\n\t\t          is_archived, archived_at, created_at, updated_at\n\t`, id, fullName, dob, guardian, memorized, levelEntry, fee, exemption, status, notes, gender, username, passwordPlain, guardianID,'
new = 'monthly_fee, fee_exemption, status, notes, gender, username, password_plain, guardian_id,\n\t\t          is_archived, archived_at, created_at, updated_at\n\t`, id, fullName, dob, guardian, memorized, levelEntry, fee, exemption, status, notes, gender, username, passwordPlain, guardianID,'
if old in content:
    content = content.replace(old, new, 1)
    open('internal/store/student.go', 'w', encoding='utf-8').write(content)
    print('OK')
else:
    print('NOT FOUND')
    idx = content.find('guardian_id      = $14')
    print(repr(content[idx:idx+300]))
