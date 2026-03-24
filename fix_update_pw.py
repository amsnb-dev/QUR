content = open('internal/store/student.go', encoding='utf-8').read()
old = '&s.MonthlyFee, &s.FeeExemption, &s.Status, &s.Notes, &s.Gender, &s.Username,\n\t\t&s.IsArchived, &s.ArchivedAt, &s.CreatedAt, &s.UpdatedAt,\n\t)\n\tif err != nil {\n\t\treturn nil, fmt.Errorf("store.UpdateStudent: %w", err)\n\t}'
new = '&s.MonthlyFee, &s.FeeExemption, &s.Status, &s.Notes, &s.Gender, &s.Username, &s.PasswordPlain,\n\t\t&s.IsArchived, &s.ArchivedAt, &s.CreatedAt, &s.UpdatedAt,\n\t)\n\tif err != nil {\n\t\treturn nil, fmt.Errorf("store.UpdateStudent: %w", err)\n\t}'
if old in content:
    content = content.replace(old, new, 1)
    open('internal/store/student.go', 'w', encoding='utf-8').write(content)
    print('OK')
else:
    print('NOT FOUND')
