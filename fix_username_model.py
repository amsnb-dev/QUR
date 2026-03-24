with open('internal/model/group_student.go', 'r', encoding='utf-8') as f:
    content = f.read()

# Add to Student struct
old1 = 'Gender         *string    `json:"gender,omitempty"`\n\tIsArchived'
new1 = 'Gender         *string    `json:"gender,omitempty"`\n\tUsername       *string    `json:"username,omitempty"`\n\tIsArchived'

# Add to CreateStudentRequest
old2 = 'Gender         *string    `json:"gender"`\n}'
new2 = 'Gender         *string    `json:"gender"`\n\tUsername       *string    `json:"username"`\n}'

for i, (old, new) in enumerate([(old1, new1), (old2, new2)]):
    if old in content:
        content = content.replace(old, new, 1)
        print(f"OK: {i+1}")
    else:
        print(f"NOT FOUND: {i+1}")

with open('internal/model/group_student.go', 'w', encoding='utf-8') as f:
    f.write(content)
