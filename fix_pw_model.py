# Fix model - add password_plain to Student struct and CreateStudentRequest
content = open('internal/model/group_student.go', encoding='utf-8').read()

old1 = 'Username       *string    `json:"username,omitempty"`\n\tIsArchived'
new1 = 'Username       *string    `json:"username,omitempty"`\n\tPasswordPlain  *string    `json:"password_plain,omitempty"`\n\tIsArchived'

old2 = 'Username       *string    `json:"username"`\n}'
new2 = 'Username       *string    `json:"username"`\n\tPasswordPlain  *string    `json:"password_plain"`\n}'

for i, (old, new) in enumerate([(old1, new1), (old2, new2)]):
    if old in content:
        content = content.replace(old, new, 1)
        print(f"model OK: {i+1}")
    else:
        print(f"model NOT FOUND: {i+1}")

open('internal/model/group_student.go', 'w', encoding='utf-8').write(content)
