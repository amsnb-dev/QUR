content = open('internal/model/group_student.go', encoding='utf-8').read()
old = 'Gender         *string    `json:"gender"`\n}\n\ntype ListStudentsFilter'
new = 'Gender         *string    `json:"gender"`\n\tUsername       *string    `json:"username"`\n\tPasswordPlain  *string    `json:"password_plain"`\n}\n\ntype ListStudentsFilter'
if old in content:
    content = content.replace(old, new, 1)
    open('internal/model/group_student.go', 'w', encoding='utf-8').write(content)
    print('OK')
else:
    print('NOT FOUND')
    idx = content.find('UpdateStudentRequest')
    print(repr(content[idx:idx+300]))
