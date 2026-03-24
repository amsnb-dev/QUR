with open('internal/model/group_student.go', 'r', encoding='utf-8') as f:
    content = f.read()
idx = content.find('type Student struct')
print(repr(content[idx:idx+500]))
