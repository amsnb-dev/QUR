content = open('internal/handler/guardian.go', encoding='utf-8').read()
old = 'schoolID := middleware.SchoolIDFrom(c)'
new = 'schoolID, _ := middleware.SchoolIDFrom(c)'
if old in content:
    content = content.replace(old, new, 1)
    open('internal/handler/guardian.go', 'w', encoding='utf-8').write(content)
    print('OK')
else:
    print('NOT FOUND')
