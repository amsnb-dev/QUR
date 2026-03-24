content = open('internal/model/group_student.go', encoding='utf-8').read()
old = 'Username       *string    `json:"username"`\n\tPasswordPlain  *string    `json:"password_plain"`\n}'
new = 'Username       *string    `json:"username"`\n\tPasswordPlain  *string    `json:"password_plain"`\n\tGuardianID     *string    `json:"guardian_id"`\n}'
if old in content:
    content = content.replace(old, new, 1)
    open('internal/model/group_student.go', 'w', encoding='utf-8').write(content)
    print('OK')
else:
    print('NOT FOUND')
