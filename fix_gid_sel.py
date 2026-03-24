content = open('internal/store/student.go', encoding='utf-8').read()

# Add to SELECT in GetStudent
old1 = 'monthly_fee, fee_exemption, status, notes, gender, username, password_plain,\n\t\t       is_archived'
new1 = 'monthly_fee, fee_exemption, status, notes, gender, username, password_plain, guardian_id,\n\t\t       is_archived'

# Add to SELECT in ListStudents
old2 = 's.monthly_fee, s.fee_exemption, s.status, s.notes, s.gender, s.username, s.password_plain,\n\t\t       s.is_archived'
new2 = 's.monthly_fee, s.fee_exemption, s.status, s.notes, s.gender, s.username, s.password_plain, s.guardian_id,\n\t\t       s.is_archived'

# Add to SCAN
old3 = '&s.Username, &s.PasswordPlain,\n\t\t\t&s.IsArchived'
new3 = '&s.Username, &s.PasswordPlain, &s.GuardianID,\n\t\t\t&s.IsArchived'

old4 = '&s.Username, &s.PasswordPlain,\n\t\t&s.IsArchived'
new4 = '&s.Username, &s.PasswordPlain, &s.GuardianID,\n\t\t&s.IsArchived'

old5 = '&s.Username, &s.PasswordPlain,\n\t\t\t\t&s.IsArchived'
new5 = '&s.Username, &s.PasswordPlain, &s.GuardianID,\n\t\t\t\t&s.IsArchived'

for i,(old,new) in enumerate([(old1,new1),(old2,new2),(old3,new3),(old4,new4),(old5,new5)]):
    c = content.count(old)
    content = content.replace(old, new)
    print(f"{'OK' if c else 'NOT FOUND'}: {i+1} ({c}x)")

open('internal/store/student.go', 'w', encoding='utf-8').write(content)
