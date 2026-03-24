content = open('internal/model/group_student.go', encoding='utf-8').read()
old = 'PasswordPlain  *string    `json:"password_plain,omitempty"`\n\tIsArchived'
new = 'PasswordPlain  *string    `json:"password_plain,omitempty"`\n\tGuardianID     *string    `json:"guardian_id,omitempty"`\n\tIsArchived'
if old in content:
    content = content.replace(old, new, 1)
    open('internal/model/group_student.go', 'w', encoding='utf-8').write(content)
    print('OK')
else:
    print('NOT FOUND')
