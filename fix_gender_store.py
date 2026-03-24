with open('internal/store/student.go', 'r', encoding='utf-8') as f:
    content = f.read()

old = '  AND ($4 = \'\' OR s.full_name ILIKE \'%\' || $4 || \'%\')\n\t`, f.IncludeArchived, f.Status, f.GroupID, f.Search).Scan(&total)'
new = '  AND ($4 = \'\' OR s.full_name ILIKE \'%\' || $4 || \'%\')\n\t\t  AND ($5 = \'\' OR s.gender = $5)\n\t`, f.IncludeArchived, f.Status, f.GroupID, f.Search, f.Gender).Scan(&total)'

old2 = '  AND ($4 = \'\' OR s.full_name ILIKE \'%\' || $4 || \'%\')\n\t\tORDER BY s.full_name\n\t\tLIMIT  $5 OFFSET $6\n\t`, f.IncludeArchived, f.Status, f.GroupID, f.Search, p.Limit, p.Offset)'
new2 = '  AND ($4 = \'\' OR s.full_name ILIKE \'%\' || $4 || \'%\')\n\t\t  AND ($5 = \'\' OR s.gender = $5)\n\t\tORDER BY s.full_name\n\t\tLIMIT  $6 OFFSET $7\n\t`, f.IncludeArchived, f.Status, f.GroupID, f.Search, f.Gender, p.Limit, p.Offset)'

for i, (old, new) in enumerate([(old, new), (old2, new2)]):
    if old in content:
        content = content.replace(old, new, 1)
        print(f"OK: {i+1}")
    else:
        print(f"NOT FOUND: {i+1}")

with open('internal/store/student.go', 'w', encoding='utf-8') as f:
    f.write(content)
