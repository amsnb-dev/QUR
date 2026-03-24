content = open('internal/store/student.go', encoding='utf-8').read()
idx = content.find('store.UpdateStudent')
print(repr(content[idx-800:idx-400]))
