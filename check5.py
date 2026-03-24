content = open('internal/store/student.go', encoding='utf-8').read()
idx = content.find('gender          = $11')
old = content[idx:idx+280]
print(repr(old))
