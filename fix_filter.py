with open('internal/model/group_student.go', 'r', encoding='utf-8') as f:
    content = f.read()

old = 'type ListStudentsFilter struct {\nIncludeArchived bool\nStatus          string\nGroupID         *uuid.UUID\nSearch          string\n}'
new = 'type ListStudentsFilter struct {\nIncludeArchived bool\nStatus          string\nGroupID         *uuid.UUID\nSearch          string\nGender          string\n}'

if old in content:
    content = content.replace(old, new)
    print("OK")
else:
    idx = content.find("ListStudentsFilter")
    print("NOT FOUND - actual:")
    print(repr(content[idx:idx+150]))

with open('internal/model/group_student.go', 'w', encoding='utf-8') as f:
    f.write(content)
