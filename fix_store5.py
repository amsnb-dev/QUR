with open('internal/store/student.go', 'r', encoding='utf-8') as f:
    content = f.read()

# Fix ListStudents scan
old1 = '&s.MonthlyFee, &s.FeeExemption, &s.Status, &s.Notes,\n\t\t\t\t&s.IsArchived, &s.ArchivedAt, &s.CreatedAt, &s.UpdatedAt,'
new1 = '&s.MonthlyFee, &s.FeeExemption, &s.Status, &s.Notes, &s.Gender,\n\t\t\t\t&s.IsArchived, &s.ArchivedAt, &s.CreatedAt, &s.UpdatedAt,'

# Fix UpdateStudent - add gender variable
old2 = '\tnotes := cur.Notes\n\tif r.Notes != nil {\n\t\tnotes = r.Notes\n\t}\n\n\tvar s model.Student'
new2 = '\tnotes := cur.Notes\n\tif r.Notes != nil {\n\t\tnotes = r.Notes\n\t}\n\tgender := cur.Gender\n\tif r.Gender != nil {\n\t\tgender = r.Gender\n\t}\n\n\tvar s model.Student'

# Fix UpdateStudent scan
old3 = '&s.MonthlyFee, &s.FeeExemption, &s.Status, &s.Notes,\n\t\t&s.IsArchived, &s.ArchivedAt, &s.CreatedAt, &s.UpdatedAt,'
new3 = '&s.MonthlyFee, &s.FeeExemption, &s.Status, &s.Notes, &s.Gender,\n\t\t&s.IsArchived, &s.ArchivedAt, &s.CreatedAt, &s.UpdatedAt,'

for i, (old, new) in enumerate([(old1, new1), (old2, new2), (old3, new3)]):
    if old in content:
        content = content.replace(old, new, 1)
        print(f"OK: {i+1}")
    else:
        print(f"NOT FOUND: {i+1}")

with open('internal/store/student.go', 'w', encoding='utf-8') as f:
    f.write(content)
