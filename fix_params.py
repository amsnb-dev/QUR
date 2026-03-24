content = open('internal/store/student.go', encoding='utf-8').read()
old = '`, id, fullName, dob, guardian, memorized, levelEntry, fee, exemption, status, notes, gender,'
new = '`, id, fullName, dob, guardian, memorized, levelEntry, fee, exemption, status, notes, gender, username, passwordPlain,'
if old in content:
    content = content.replace(old, new, 1)
    open('internal/store/student.go', 'w', encoding='utf-8').write(content)
    print('OK')
else:
    print('NOT FOUND')
