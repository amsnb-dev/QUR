content = open('internal/store/student.go', encoding='utf-8').read()

# Fix SET to add username and password_plain
old = '       gender          = $11\n\t\tWHERE  id = $1'
new = '       gender          = $11,\n\t\t       username         = $12,\n\t\t       password_plain   = $13\n\t\tWHERE  id = $1'

# Fix params to add username and password_plain
old2 = '`, id, fullName, dob, guardian, memorized, levelEntry, fee, exemption, status, notes, gender, cur.Username,'
new2 = '`, id, fullName, dob, guardian, memorized, levelEntry, fee, exemption, status, notes, gender, username, passwordPlain,'

# Fix UpdateStudent to add username and passwordPlain variables
old3 = '\tgender := cur.Gender\n\tif r.Gender != nil {\n\t\tgender = r.Gender\n\t}\n\n\tvar s model.Student'
new3 = '\tgender := cur.Gender\n\tif r.Gender != nil {\n\t\tgender = r.Gender\n\t}\n\tusername := cur.Username\n\tif r.Username != nil {\n\t\tusername = r.Username\n\t}\n\tpasswordPlain := cur.PasswordPlain\n\tif r.PasswordPlain != nil {\n\t\tpasswordPlain = r.PasswordPlain\n\t}\n\n\tvar s model.Student'

for i, (old, new) in enumerate([(old, new), (old2, new2), (old3, new3)]):
    if old in content:
        content = content.replace(old, new, 1)
        print(f"OK: {i+1}")
    else:
        print(f"NOT FOUND: {i+1}")

open('internal/store/student.go', 'w', encoding='utf-8').write(content)
