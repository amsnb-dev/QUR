content = open('internal/model/guardian.go', encoding='utf-8').read()
old = 'IsActive      *bool   `json:"is_active"`\n}'
new = 'Relation      *string `json:"relation"`\n\tIsActive      *bool   `json:"is_active"`\n}'
if old in content:
    content = content.replace(old, new, 1)
    open('internal/model/guardian.go', 'w', encoding='utf-8').write(content)
    print('OK')
else:
    print('NOT FOUND')
    idx = content.find('IsActive')
    print(repr(content[idx-50:idx+80]))
