package store

import (
	"fmt"
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/quran-school/api/internal/model"
)

// ── Subjects ───────────────────────────────────────────────

func CreateSubject(ctx context.Context, tx pgx.Tx, schoolID uuid.UUID, createdBy uuid.UUID, req *model.CreateSubjectRequest) (*model.Subject, error) {
	color := "#1a7a4e"
	if req.Color != nil {
		color = *req.Color
	}
	icon := "📖"
	if req.Icon != nil {
		icon = *req.Icon
	}
	sortOrder := 0
	if req.SortOrder != nil {
		sortOrder = *req.SortOrder
	}

	var s model.Subject
	err := tx.QueryRow(ctx, `
		INSERT INTO subjects (school_id, name, description, category, color, icon, sort_order, created_by)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
		RETURNING id,school_id,name,description,category,color,icon,is_active,sort_order,created_by,created_at,updated_at
	`, schoolID, req.Name, req.Description, req.Category, color, icon, sortOrder, createdBy).Scan(
		&s.ID, &s.SchoolID, &s.Name, &s.Description, &s.Category,
		&s.Color, &s.Icon, &s.IsActive, &s.SortOrder, &s.CreatedBy,
		&s.CreatedAt, &s.UpdatedAt,
	)
	return &s, err
}

func ListSubjects(ctx context.Context, tx pgx.Tx, includeInactive bool) ([]model.Subject, error) {
	q := `
		SELECT s.id, s.school_id, s.name, s.description, s.category, s.color, s.icon,
		       s.is_active, s.sort_order, s.created_by, s.created_at, s.updated_at,
		       COUNT(sl.id) AS levels_count
		FROM subjects s
		LEFT JOIN subject_levels sl ON sl.subject_id=s.id AND sl.school_id=s.school_id AND sl.is_active=TRUE
	`
	if !includeInactive {
		q += ` WHERE s.is_active=TRUE `
	}
	q += ` GROUP BY s.id ORDER BY s.sort_order, s.name`

	rows, err := tx.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []model.Subject
	for rows.Next() {
		var s model.Subject
		if err := rows.Scan(
			&s.ID, &s.SchoolID, &s.Name, &s.Description, &s.Category,
			&s.Color, &s.Icon, &s.IsActive, &s.SortOrder, &s.CreatedBy,
			&s.CreatedAt, &s.UpdatedAt, &s.LevelsCount,
		); err != nil {
			return nil, err
		}
		list = append(list, s)
	}
	return list, nil
}

func GetSubject(ctx context.Context, tx pgx.Tx, id uuid.UUID) (*model.Subject, error) {
	var s model.Subject
	err := tx.QueryRow(ctx, `
		SELECT id,school_id,name,description,category,color,icon,is_active,sort_order,created_by,created_at,updated_at
		FROM subjects WHERE id=$1
	`, id).Scan(
		&s.ID, &s.SchoolID, &s.Name, &s.Description, &s.Category,
		&s.Color, &s.Icon, &s.IsActive, &s.SortOrder, &s.CreatedBy,
		&s.CreatedAt, &s.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	return &s, err
}

func UpdateSubject(ctx context.Context, tx pgx.Tx, id uuid.UUID, req *model.UpdateSubjectRequest) (*model.Subject, error) {
	var s model.Subject
	err := tx.QueryRow(ctx, `
		UPDATE subjects SET
			name        = COALESCE($2, name),
			description = COALESCE($3, description),
			category    = COALESCE($4, category),
			color       = COALESCE($5, color),
			icon        = COALESCE($6, icon),
			is_active   = COALESCE($7, is_active),
			sort_order  = COALESCE($8, sort_order)
		WHERE id=$1
		RETURNING id,school_id,name,description,category,color,icon,is_active,sort_order,created_by,created_at,updated_at
	`, id, req.Name, req.Description, req.Category, req.Color, req.Icon, req.IsActive, req.SortOrder).Scan(
		&s.ID, &s.SchoolID, &s.Name, &s.Description, &s.Category,
		&s.Color, &s.Icon, &s.IsActive, &s.SortOrder, &s.CreatedBy,
		&s.CreatedAt, &s.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	return &s, err
}

// ── SubjectLevels ──────────────────────────────────────────

func CreateSubjectLevel(ctx context.Context, tx pgx.Tx, schoolID uuid.UUID, subjectID uuid.UUID, req *model.CreateSubjectLevelRequest) (*model.SubjectLevel, error) {
	orderIndex := 0
	if req.OrderIndex != nil {
		orderIndex = *req.OrderIndex
	}
	var l model.SubjectLevel
	err := tx.QueryRow(ctx, `
		INSERT INTO subject_levels (school_id, subject_id, name, description, order_index, criteria)
		VALUES ($1,$2,$3,$4,$5,$6)
		RETURNING id,school_id,subject_id,name,description,order_index,criteria,is_active,created_at,updated_at
	`, schoolID, subjectID, req.Name, req.Description, orderIndex, req.Criteria).Scan(
		&l.ID, &l.SchoolID, &l.SubjectID, &l.Name, &l.Description,
		&l.OrderIndex, &l.Criteria, &l.IsActive, &l.CreatedAt, &l.UpdatedAt,
	)
	return &l, err
}

func ListSubjectLevels(ctx context.Context, tx pgx.Tx, subjectID uuid.UUID) ([]model.SubjectLevel, error) {
	rows, err := tx.Query(ctx, `
		SELECT l.id, l.school_id, l.subject_id, l.name, l.description, l.order_index,
		       l.criteria, l.is_active, l.created_at, l.updated_at, s.name
		FROM subject_levels l
		JOIN subjects s ON s.id=l.subject_id
		WHERE l.subject_id=$1 AND l.is_active=TRUE
		ORDER BY l.order_index, l.name
	`, subjectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []model.SubjectLevel
	for rows.Next() {
		var l model.SubjectLevel
		if err := rows.Scan(
			&l.ID, &l.SchoolID, &l.SubjectID, &l.Name, &l.Description,
			&l.OrderIndex, &l.Criteria, &l.IsActive, &l.CreatedAt, &l.UpdatedAt,
			&l.SubjectName,
		); err != nil {
			return nil, err
		}
		list = append(list, l)
	}
	return list, nil
}

func UpdateSubjectLevel(ctx context.Context, tx pgx.Tx, id uuid.UUID, req *model.UpdateSubjectLevelRequest) (*model.SubjectLevel, error) {
	var l model.SubjectLevel
	err := tx.QueryRow(ctx, `
		UPDATE subject_levels SET
			name        = COALESCE($2, name),
			description = COALESCE($3, description),
			order_index = COALESCE($4, order_index),
			criteria    = COALESCE($5, criteria),
			is_active   = COALESCE($6, is_active)
		WHERE id=$1
		RETURNING id,school_id,subject_id,name,description,order_index,criteria,is_active,created_at,updated_at
	`, id, req.Name, req.Description, req.OrderIndex, req.Criteria, req.IsActive).Scan(
		&l.ID, &l.SchoolID, &l.SubjectID, &l.Name, &l.Description,
		&l.OrderIndex, &l.Criteria, &l.IsActive, &l.CreatedAt, &l.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	return &l, err
}

// ── StudentSubjects ────────────────────────────────────────

func AssignSubjectToStudent(ctx context.Context, tx pgx.Tx, schoolID uuid.UUID, studentID uuid.UUID, assignedBy uuid.UUID, req *model.AssignSubjectRequest) (*model.StudentSubject, error) {
	startedAt := time.Now().Format("2006-01-02")
	if req.StartedAt != nil {
		startedAt = *req.StartedAt
	}

	var ss model.StudentSubject
	err := tx.QueryRow(ctx, `
		INSERT INTO student_subjects (school_id, student_id, subject_id, current_level_id, started_at, notes, assigned_by)
		VALUES ($1,$2,$3,$4,$5,$6,$7)
		ON CONFLICT (school_id, student_id, subject_id) DO UPDATE SET
			current_level_id = COALESCE(EXCLUDED.current_level_id, student_subjects.current_level_id),
			status = 'active',
			notes = COALESCE(EXCLUDED.notes, student_subjects.notes)
		RETURNING id,school_id,student_id,subject_id,current_level_id,status,started_at::TEXT,notes,assigned_by,created_at,updated_at
	`, schoolID, studentID, req.SubjectID, req.CurrentLevelID, startedAt, req.Notes, assignedBy).Scan(
		&ss.ID, &ss.SchoolID, &ss.StudentID, &ss.SubjectID, &ss.CurrentLevelID,
		&ss.Status, &ss.StartedAt, &ss.Notes, &ss.AssignedBy, &ss.CreatedAt, &ss.UpdatedAt,
	)
	return &ss, err
}

func ListStudentSubjects(ctx context.Context, tx pgx.Tx, studentID uuid.UUID) ([]model.StudentSubject, error) {
	rows, err := tx.Query(ctx, `
		SELECT ss.id, ss.school_id, ss.student_id, ss.subject_id, ss.current_level_id,
		       ss.status, ss.started_at::TEXT, ss.notes, ss.assigned_by, ss.created_at, ss.updated_at,
		       s.name, s.icon, s.color,
		       sl.name
		FROM student_subjects ss
		JOIN subjects s ON s.id=ss.subject_id
		LEFT JOIN subject_levels sl ON sl.id=ss.current_level_id
		WHERE ss.student_id=$1
		ORDER BY s.sort_order, s.name
	`, studentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []model.StudentSubject
	for rows.Next() {
		var ss model.StudentSubject
		if err := rows.Scan(
			&ss.ID, &ss.SchoolID, &ss.StudentID, &ss.SubjectID, &ss.CurrentLevelID,
			&ss.Status, &ss.StartedAt, &ss.Notes, &ss.AssignedBy, &ss.CreatedAt, &ss.UpdatedAt,
			&ss.SubjectName, &ss.SubjectIcon, &ss.SubjectColor,
			&ss.CurrentLevelName,
		); err != nil {
			return nil, err
		}
		list = append(list, ss)
	}
	return list, nil
}

func UpdateStudentSubject(ctx context.Context, tx pgx.Tx, id uuid.UUID, req *model.UpdateStudentSubjectRequest) (*model.StudentSubject, error) {
	var ss model.StudentSubject
	err := tx.QueryRow(ctx, `
		UPDATE student_subjects SET
			current_level_id = COALESCE($2, current_level_id),
			status           = COALESCE($3, status),
			notes            = COALESCE($4, notes)
		WHERE id=$1
		RETURNING id,school_id,student_id,subject_id,current_level_id,status,started_at::TEXT,notes,assigned_by,created_at,updated_at
	`, id, req.CurrentLevelID, req.Status, req.Notes).Scan(
		&ss.ID, &ss.SchoolID, &ss.StudentID, &ss.SubjectID, &ss.CurrentLevelID,
		&ss.Status, &ss.StartedAt, &ss.Notes, &ss.AssignedBy, &ss.CreatedAt, &ss.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	return &ss, err
}

// ── SubjectSessions ────────────────────────────────────────

func CreateSubjectSession(ctx context.Context, tx pgx.Tx, schoolID uuid.UUID, recordedBy uuid.UUID, req *model.CreateSubjectSessionRequest) (*model.SubjectSession, error) {
	sessionDate := time.Now().Format("2006-01-02")
	if req.SessionDate != "" {
		sessionDate = req.SessionDate
	}

	var s model.SubjectSession
	err := tx.QueryRow(ctx, `
		INSERT INTO subject_sessions
		  (school_id, student_id, subject_id, teacher_id, session_date, content,
		   pages_count, duration_minutes, performance, score, level_id, notes, recorded_by)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)
		RETURNING id,school_id,student_id,subject_id,teacher_id,session_date::TEXT,
		          content,pages_count,duration_minutes,performance,score,level_id,notes,recorded_by,created_at,updated_at
	`, schoolID, req.StudentID, req.SubjectID, req.TeacherID, sessionDate,
		req.Content, req.PagesCount, req.DurationMinutes, req.Performance,
		req.Score, req.LevelID, req.Notes, recordedBy).Scan(
		&s.ID, &s.SchoolID, &s.StudentID, &s.SubjectID, &s.TeacherID,
		&s.SessionDate, &s.Content, &s.PagesCount, &s.DurationMinutes,
		&s.Performance, &s.Score, &s.LevelID, &s.Notes, &s.RecordedBy,
		&s.CreatedAt, &s.UpdatedAt,
	)
	return &s, err
}

func ListSubjectSessions(ctx context.Context, tx pgx.Tx, f model.ListSubjectSessionsFilter, p Page) ([]model.SubjectSession, int, error) {
	args := []any{}
	where := "WHERE 1=1 "
	n := 1

	if f.StudentID != nil {
		where += fmt.Sprintf(" AND ss.student_id=$%d", n)
		args = append(args, *f.StudentID)
		n++
	}
	if f.SubjectID != nil {
		where += fmt.Sprintf(" AND ss.subject_id=$%d", n)
		args = append(args, *f.SubjectID)
		n++
	}
	if f.TeacherID != nil {
		where += fmt.Sprintf(" AND ss.teacher_id=$%d", n)
		args = append(args, *f.TeacherID)
		n++
	}
	if f.DateFrom != nil {
		where += fmt.Sprintf(" AND ss.session_date>=$%d", n)
		args = append(args, *f.DateFrom)
		n++
	}
	if f.DateTo != nil {
		where += fmt.Sprintf(" AND ss.session_date<=$%d", n)
		args = append(args, *f.DateTo)
		n++
	}

	countArgs := make([]any, len(args))
	copy(countArgs, args)

	var total int
	if err := tx.QueryRow(ctx, "SELECT COUNT(*) FROM subject_sessions ss "+where, countArgs...).Scan(&total); err != nil {
		return nil, 0, err
	}

	args = append(args, p.Limit, p.Offset)
	rows, err := tx.Query(ctx, `
		SELECT ss.id, ss.school_id, ss.student_id, ss.subject_id, ss.teacher_id,
		       ss.session_date::TEXT, ss.content, ss.pages_count, ss.duration_minutes,
		       ss.performance, ss.score, ss.level_id, ss.notes, ss.recorded_by,
		       ss.created_at, ss.updated_at,
		       st.full_name, s.name,
		       t.full_name,
		       sl.name
		FROM subject_sessions ss
		JOIN students st ON st.id=ss.student_id
		JOIN subjects s  ON s.id=ss.subject_id
		LEFT JOIN teachers t  ON t.id=ss.teacher_id
		LEFT JOIN subject_levels sl ON sl.id=ss.level_id
		`+where+`
		ORDER BY ss.session_date DESC, ss.created_at DESC
		LIMIT $`+fmt.Sprintf("%d", n)+` OFFSET $`+fmt.Sprintf("%d", n+1),
		args...,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var list []model.SubjectSession
	for rows.Next() {
		var s model.SubjectSession
		if err := rows.Scan(
			&s.ID, &s.SchoolID, &s.StudentID, &s.SubjectID, &s.TeacherID,
			&s.SessionDate, &s.Content, &s.PagesCount, &s.DurationMinutes,
			&s.Performance, &s.Score, &s.LevelID, &s.Notes, &s.RecordedBy,
			&s.CreatedAt, &s.UpdatedAt,
			&s.StudentName, &s.SubjectName, &s.TeacherName, &s.LevelName,
		); err != nil {
			return nil, 0, err
		}
		list = append(list, s)
	}
	return list, total, nil
}
