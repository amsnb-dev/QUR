content = open('internal/store/student.go', encoding='utf-8').read()

# Add guardian_id to UPDATE SET
old1 = '       gender          = $11,\n\t\t       username         = $12,\n\t\t       password_plain   = $13\n\t\tWHERE  id = $1'
new1 = '       gender          = $11,\n\t\t       username         = $12,\n\t\t       password_plain   = $13,\n\t\t       guardian_id      = $14\n\t\tWHERE  id = $1'

# Add guardian_id to UPDATE params
old2 = '`, id, fullName, dob, guardian, memorized, levelEntry, fee, exemption, status, notes, gender, username, passwordPlain,'
new2 = '`, id, fullName, dob, guardian, memorized, levelEntry, fee, exemption, status, notes, gender, username, passwordPlain, guardianID,'

# Add guardian_id variable
old3 = '\tgender := cur.Gender\n\tif r.Gender != nil {\n\t\tgender = r.Gender\n\t}\n\tusername'
new3 = '\tgender := cur.Gender\n\tif r.Gender != nil {\n\t\tgender = r.Gender\n\t}\n\tguardianID := cur.GuardianID\n\tif r.GuardianID != nil {\n\t\tguardianID = r.GuardianID\n\t}\n\tusername'

for i,(old,new) in enumerate([(old1,new1),(old2,new2),(old3,new3)]):
    if old in content:
        content = content.replace(old, new, 1)
        print(f"OK: {i+1}")
    else:
        print(f"NOT FOUND: {i+1}")

open('internal/store/student.go', 'w', encoding='utf-8').write(content)
