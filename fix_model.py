with open('internal/model/group_student.go', 'r', encoding='utf-8') as f:
    content = f.read()

# Add Gender to CreateStudentRequest
old1 = 'Notes          *string    `json:"notes"`\n}'
new1 = 'Notes          *string    `json:"notes"`\n\tGender         *string    `json:"gender"`\n}'

# Add Gender to UpdateStudentRequest  
old2 = 'Notes          *string    `json:"notes"`\n}\n\ntype ListStudentsFilter'
new2 = 'Notes          *string    `json:"notes"`\n\tGender         *string    `json:"gender"`\n}\n\ntype ListStudentsFilter'

if old1 in content:
    content = content.replace(old1, new1, 1)
    print("Create OK")
else:
    print("Create NOT FOUND")

if old2 in content:
    content = content.replace(old2, new2, 1)
    print("Update OK")
else:
    print("Update NOT FOUND")

with open('internal/model/group_student.go', 'w', encoding='utf-8') as f:
    f.write(content)
