content = open('internal/model/group_student.go', encoding='utf-8').read()

# Add to Student struct
old1 = 'Username       *string    `json:"username,omitempty"`\n\tPasswordPlain  *string    `json:"password_plain,omitempty"`\n\tGender'
new1 = 'Username       *string    `json:"username,omitempty"`\n\tPasswordPlain  *string    `json:"password_plain,omitempty"`\n\tGuardianID     *string    `json:"guardian_id,omitempty"`\n\tGender'

# Add to UpdateStudentRequest
old2 = 'Username       *string    `json:"username"`\n\tPasswordPlain  *string    `json:"password_plain"`\n\tGuardianID     *string    `json:"guardian_id"`\n}'
# Already exists from earlier - check
old2b = 'Username       *string    `json:"username"`\n\tPasswordPlain  *string    `json:"password_plain"`\n}'
new2b = 'Username       *string    `json:"username"`\n\tPasswordPlain  *string    `json:"password_plain"`\n\tGuardianID     *string    `json:"guardian_id"`\n}'

for i,(old,new) in enumerate([(old1,new1),(old2b,new2b)]):
    if old in content:
        content = content.replace(old, new, 1)
        print(f"OK: {i+1}")
    else:
        print(f"NOT FOUND: {i+1}")

open('internal/model/group_student.go', 'w', encoding='utf-8').write(content)
