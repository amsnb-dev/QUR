with open('internal/store/student.go', 'r', encoding='utf-8') as f:
    content = f.read()

old = 'func ListStudents(ctx context.Context, tx pgx.Tx, f model.ListStudentsFilter, p Page) ([]model.Student, int, error) {\n\tvar total int\n\tif err := tx.QueryRow(ctx, `\n\t\tSELECT COUNT(*)\n\t\tFROM   students\n\t\tWHERE ($1 OR is_archived = FALSE)\n\t\t  AND ($2 = \'\' OR status = $2)\n\t`, f.IncludeArchived, f.Status).Scan(&total); err != nil {\n\t\treturn nil, 0, fmt.Errorf("store.ListStudents count: %w", err)\n\t}\n\n\trows, err := tx.Query(ctx, `\n\t\tSELECT id, school_id, full_name, date_of_birth, guardian_name,\n\t\t       enrollment_date, memorized_parts, level_on_entry,\n\t\t       monthly_fee, fee_exemption, status, notes,\n\t\t       is_archived, archived_at, created_at, updated_at\n\t\tFROM   students\n\t\tWHERE ($1 OR is_archived = FALSE)\n\t\t  AND ($2 = \'\' OR status = $2)\n\t\tORDER BY full_name\n\t\tLIMIT  $3 OFFSET $4\n\t`, f.IncludeArchived, f.Status, p.Limit, p.Offset)'

new = 'func ListStudents(ctx context.Context, tx pgx.Tx, f model.ListStudentsFilter, p Page) ([]model.Student, int, error) {\n\tvar total int\n\tif err := tx.QueryRow(ctx, `\n\t\tSELECT COUNT(*)\n\t\tFROM   students s\n\t\tWHERE ($1 OR s.is_archived = FALSE)\n\t\t  AND ($2 = \'\' OR s.status = $2)\n\t\t  AND ($3::uuid IS NULL OR EXISTS (\n\t\t        SELECT 1 FROM student_groups sg\n\t\t        WHERE sg.student_id = s.id AND sg.group_id = $3 AND sg.end_date IS NULL\n\t\t  ))\n\t\t  AND ($4 = \'\' OR s.full_name ILIKE \'%\' || $4 || \'%\')\n\t`, f.IncludeArchived, f.Status, f.GroupID, f.Search).Scan(&total); err != nil {\n\t\treturn nil, 0, fmt.Errorf("store.ListStudents count: %w", err)\n\t}\n\n\trows, err := tx.Query(ctx, `\n\t\tSELECT s.id, s.school_id, s.full_name, s.date_of_birth, s.guardian_name,\n\t\t       s.enrollment_date, s.memorized_parts, s.level_on_entry,\n\t\t       s.monthly_fee, s.fee_exemption, s.status, s.notes,\n\t\t       s.is_archived, s.archived_at, s.created_at, s.updated_at\n\t\tFROM   students s\n\t\tWHERE ($1 OR s.is_archived = FALSE)\n\t\t  AND ($2 = \'\' OR s.status = $2)\n\t\t  AND ($3::uuid IS NULL OR EXISTS (\n\t\t        SELECT 1 FROM student_groups sg\n\t\t        WHERE sg.student_id = s.id AND sg.group_id = $3 AND sg.end_date IS NULL\n\t\t  ))\n\t\t  AND ($4 = \'\' OR s.full_name ILIKE \'%\' || $4 || \'%\')\n\t\tORDER BY s.full_name\n\t\tLIMIT  $5 OFFSET $6\n\t`, f.IncludeArchived, f.Status, f.GroupID, f.Search, p.Limit, p.Offset)'

if old in content:
    content = content.replace(old, new)
    with open('internal/store/student.go', 'w', encoding='utf-8') as f:
        f.write(content)
    print("SUCCESS")
else:
    print("NOT FOUND")
