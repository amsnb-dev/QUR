package store

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/quran-school/api/internal/model"
)

// â”€â”€ Students â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€

func CreateStudent(ctx context.Context, tx pgx.Tx, schoolID uuid.UUID, r *model.CreateStudentRequest) (*model.Student, error) {
	enrollDate := time.Now().UTC().Truncate(24 * time.Hour)
	if r.EnrollmentDate != nil {
		enrollDate = *r.EnrollmentDate
	}
	memorized := 0.0
	if r.MemorizedParts != nil {
		memorized = *r.MemorizedParts
	}
	exemption := "none"
	if r.FeeExemption != nil {
		exemption = *r.FeeExemption
	}

	var s model.Student
	err := tx.QueryRow(ctx, `
		INSERT INTO students
		       (school_id, full_name, date_of_birth, guardian_name,
		        enrollment_date, memorized_parts, level_on_entry,
		        monthly_fee, fee_exemption, notes, gender, username, password_plain, guardian_id)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14)
		RETURNING id, school_id, full_name, date_of_birth, guardian_name,
		          enrollment_date, memorized_parts, level_on_entry,
		          monthly_fee, fee_exemption, status, notes, gender, username, password_plain, guardian_id,
		          is_archived, archived_at, created_at, updated_at
	`, schoolID, r.FullName, r.DateOfBirth, r.GuardianName,
		enrollDate, memorized, r.LevelOnEntry,
		r.MonthlyFee, exemption, r.Notes, r.Gender, r.Username, r.PasswordPlain, r.GuardianID,
	).Scan(
		&s.ID, &s.SchoolID, &s.FullName, &s.DateOfBirth, &s.GuardianName,
		&s.EnrollmentDate, &s.MemorizedParts, &s.LevelOnEntry,
		&s.MonthlyFee, &s.FeeExemption, &s.Status, &s.Notes, &s.Gender, &s.Username, &s.PasswordPlain, &s.GuardianID,
		&s.IsArchived, &s.ArchivedAt, &s.CreatedAt, &s.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("store.CreateStudent: %w", err)
	}
	return &s, nil
}

// ListStudents returns a paginated list of students.
func ListStudents(ctx context.Context, tx pgx.Tx, f model.ListStudentsFilter, p Page) ([]model.Student, int, error) {
	var total int
	if err := tx.QueryRow(ctx, `
		SELECT COUNT(*)
		FROM   students s
		WHERE ($1 OR s.is_archived = FALSE)
		  AND ($2 = '' OR s.status = $2)
		  AND ($3::uuid IS NULL OR EXISTS (
		        SELECT 1 FROM student_groups sg
		        WHERE sg.student_id = s.id AND sg.group_id = $3 AND sg.end_date IS NULL
		  ))
		  AND ($4 = '' OR s.full_name ILIKE '%' || $4 || '%')
		  AND ($5 = '' OR s.gender = $5)
	`, f.IncludeArchived, f.Status, f.GroupID, f.Search, f.Gender).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("store.ListStudents count: %w", err)
	}

	rows, err := tx.Query(ctx, `
		SELECT s.id, s.school_id, s.full_name, s.date_of_birth, s.guardian_name,
		       s.enrollment_date, s.memorized_parts, s.level_on_entry,
		       s.monthly_fee, s.fee_exemption, s.status, s.notes, s.gender, s.username, s.password_plain, s.guardian_id,
		       s.is_archived, s.archived_at, s.created_at, s.updated_at
		FROM   students s
		WHERE ($1 OR s.is_archived = FALSE)
		  AND ($2 = '' OR s.status = $2)
		  AND ($3::uuid IS NULL OR EXISTS (
		        SELECT 1 FROM student_groups sg
		        WHERE sg.student_id = s.id AND sg.group_id = $3 AND sg.end_date IS NULL
		  ))
		  AND ($4 = '' OR s.full_name ILIKE '%' || $4 || '%')
		  AND ($5 = '' OR s.gender = $5)
		ORDER BY s.full_name
		LIMIT  $6 OFFSET $7
	`, f.IncludeArchived, f.Status, f.GroupID, f.Search, f.Gender, p.Limit, p.Offset)
	if err != nil {
		return nil, 0, fmt.Errorf("store.ListStudents: %w", err)
	}
	defer rows.Close()

	var list []model.Student
	for rows.Next() {
		var s model.Student
		if err := rows.Scan(
			&s.ID, &s.SchoolID, &s.FullName, &s.DateOfBirth, &s.GuardianName,
			&s.EnrollmentDate, &s.MemorizedParts, &s.LevelOnEntry,
			&s.MonthlyFee, &s.FeeExemption, &s.Status, &s.Notes, &s.Gender, &s.Username, &s.PasswordPlain, &s.GuardianID,
			&s.IsArchived, &s.ArchivedAt, &s.CreatedAt, &s.UpdatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("store.ListStudents scan: %w", err)
		}
		list = append(list, s)
	}
	return list, total, rows.Err()
}

func GetStudent(ctx context.Context, tx pgx.Tx, id uuid.UUID) (*model.Student, error) {
	var s model.Student
	err := tx.QueryRow(ctx, `
		SELECT id, school_id, full_name, date_of_birth, guardian_name,
		       enrollment_date, memorized_parts, level_on_entry,
		       monthly_fee, fee_exemption, status, notes, gender, username, password_plain, guardian_id,
		       is_archived, archived_at, created_at, updated_at
		FROM   students
		WHERE  id = $1
	`, id).Scan(
		&s.ID, &s.SchoolID, &s.FullName, &s.DateOfBirth, &s.GuardianName,
		&s.EnrollmentDate, &s.MemorizedParts, &s.LevelOnEntry,
		&s.MonthlyFee, &s.FeeExemption, &s.Status, &s.Notes, &s.Gender, &s.Username, &s.PasswordPlain, &s.GuardianID,
		&s.IsArchived, &s.ArchivedAt, &s.CreatedAt, &s.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("store.GetStudent: %w", err)
	}
	return &s, nil
}

func UpdateStudent(ctx context.Context, tx pgx.Tx, id uuid.UUID, r *model.UpdateStudentRequest) (*model.Student, error) {
	cur, err := GetStudent(ctx, tx, id)
	if err != nil || cur == nil {
		return cur, err
	}

	fullName := cur.FullName
	if r.FullName != nil {
		fullName = *r.FullName
	}
	dob := cur.DateOfBirth
	if r.DateOfBirth != nil {
		dob = r.DateOfBirth
	}
	guardian := cur.GuardianName
	if r.GuardianName != nil {
		guardian = r.GuardianName
	}
	memorized := cur.MemorizedParts
	if r.MemorizedParts != nil {
		memorized = *r.MemorizedParts
	}
	levelEntry := cur.LevelOnEntry
	if r.LevelOnEntry != nil {
		levelEntry = r.LevelOnEntry
	}
	fee := cur.MonthlyFee
	if r.MonthlyFee != nil {
		fee = r.MonthlyFee
	}
	exemption := cur.FeeExemption
	if r.FeeExemption != nil {
		exemption = *r.FeeExemption
	}
	status := cur.Status
	if r.Status != nil {
		status = *r.Status
	}
	notes := cur.Notes
	if r.Notes != nil {
		notes = r.Notes
	}
	gender := cur.Gender
	if r.Gender != nil {
		gender = r.Gender
	}
	guardianID := cur.GuardianID
	if r.GuardianID != nil {
		guardianID = r.GuardianID
	}
	username := cur.Username
	if r.Username != nil {
		username = r.Username
	}
	passwordPlain := cur.PasswordPlain
	if r.PasswordPlain != nil {
		passwordPlain = r.PasswordPlain
	}

	var s model.Student
	err = tx.QueryRow(ctx, `
		UPDATE students
		SET    full_name       = $2,
		       date_of_birth   = $3,
		       guardian_name   = $4,
		       memorized_parts = $5,
		       level_on_entry  = $6,
		       monthly_fee     = $7,
		       fee_exemption   = $8,
		       status          = $9,
		       notes           = $10,
		       gender          = $11,
		       username         = $12,
		       password_plain   = $13,
		       guardian_id      = $14
		WHERE  id = $1
		RETURNING id, school_id, full_name, date_of_birth, guardian_name,
		          enrollment_date, memorized_parts, level_on_entry,
		          monthly_fee, fee_exemption, status, notes, gender, username, password_plain, guardian_id,
		          is_archived, archived_at, created_at, updated_at
	`, id, fullName, dob, guardian, memorized, levelEntry, fee, exemption, status, notes, gender, username, passwordPlain, guardianID,
	).Scan(
		&s.ID, &s.SchoolID, &s.FullName, &s.DateOfBirth, &s.GuardianName,
		&s.EnrollmentDate, &s.MemorizedParts, &s.LevelOnEntry,
		&s.MonthlyFee, &s.FeeExemption, &s.Status, &s.Notes, &s.Gender, &s.Username, &s.PasswordPlain, &s.GuardianID,
		&s.IsArchived, &s.ArchivedAt, &s.CreatedAt, &s.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("store.UpdateStudent: %w", err)
	}
	return &s, nil
}

// ArchiveStudent calls the DB function archive_student() which enforces
// business rules (no open groups, no unpaid invoices) then writes to audit_logs.
func ArchiveStudent(ctx context.Context, tx pgx.Tx, studentID, schoolID, archivedBy uuid.UUID) error {
	// The PG function already writes to audit_logs internally, so we only
	// call it here â€” no double-audit needed.
	_, err := tx.Exec(ctx,
		"SELECT archive_student($1, $2, $3)",
		studentID, schoolID, archivedBy,
	)
	if err != nil {
		return fmt.Errorf("store.ArchiveStudent: %w", err)
	}
	return nil
}

// â”€â”€ StudentGroups â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€

// AddStudentGroup adds a student to a group.
// If is_primary=true and an open primary exists, it closes it first (auto-rotate).
// Returns ErrConflict if the student is already enrolled in that group (open).
func AddStudentGroup(ctx context.Context, tx pgx.Tx, schoolID, studentID uuid.UUID, r *model.AddStudentGroupRequest, actorID uuid.UUID) (*model.StudentGroup, error) {
	isPrimary := true
	if r.IsPrimary != nil {
		isPrimary = *r.IsPrimary
	}
	startDate := time.Now().UTC().Truncate(24 * time.Hour)
	if r.StartDate != nil {
		startDate = *r.StartDate
	}

	var exists bool
	if err := tx.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM student_groups
			WHERE  school_id  = $1
			  AND  student_id = $2
			  AND  group_id   = $3
			  AND  end_date IS NULL
		)
	`, schoolID, studentID, r.GroupID).Scan(&exists); err != nil {
		return nil, fmt.Errorf("store.AddStudentGroup dup check: %w", err)
	}
	if exists {
		return nil, ErrConflict
	}

	if isPrimary {
		if _, err := tx.Exec(ctx, `
			UPDATE student_groups
			SET    end_date = $3
			WHERE  school_id  = $1
			  AND  student_id = $2
			  AND  is_primary = TRUE
			  AND  end_date IS NULL
		`, schoolID, studentID, startDate); err != nil {
			return nil, fmt.Errorf("store.AddStudentGroup close old primary: %w", err)
		}
	}

	var sg model.StudentGroup
	err := tx.QueryRow(ctx, `
		INSERT INTO student_groups (school_id, student_id, group_id, start_date, is_primary)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, school_id, student_id, group_id, start_date, end_date, is_primary, created_at
	`, schoolID, studentID, r.GroupID, startDate, isPrimary,
	).Scan(
		&sg.ID, &sg.SchoolID, &sg.StudentID, &sg.GroupID,
		&sg.StartDate, &sg.EndDate, &sg.IsPrimary, &sg.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("store.AddStudentGroup insert: %w", err)
	}
	sg.IsActive = sg.EndDate == nil

	_ = InsertAudit(ctx, tx, AuditParams{
		SchoolID:  &schoolID,
		UserID:    &actorID,
		Action:    "INSERT",
		TableName: "student_groups",
		RecordID:  sg.ID.String(),
		NewValues: map[string]any{
			"student_id": studentID,
			"group_id":   r.GroupID,
			"is_primary": isPrimary,
		},
	})
	return &sg, nil
}

// CloseStudentGroup sets end_date on a membership.
func CloseStudentGroup(ctx context.Context, tx pgx.Tx, schoolID, membershipID uuid.UUID, endDate time.Time, actorID uuid.UUID) (*model.StudentGroup, error) {
	var sg model.StudentGroup
	err := tx.QueryRow(ctx, `
		UPDATE student_groups
		SET    end_date = $3
		WHERE  id        = $2
		  AND  school_id = $1
		  AND  end_date IS NULL
		RETURNING id, school_id, student_id, group_id, start_date, end_date, is_primary, created_at
	`, schoolID, membershipID, endDate).Scan(
		&sg.ID, &sg.SchoolID, &sg.StudentID, &sg.GroupID,
		&sg.StartDate, &sg.EndDate, &sg.IsPrimary, &sg.CreatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("store.CloseStudentGroup: %w", err)
	}
	sg.IsActive = false

	_ = InsertAudit(ctx, tx, AuditParams{
		SchoolID:  &schoolID,
		UserID:    &actorID,
		Action:    "CLOSE",
		TableName: "student_groups",
		RecordID:  membershipID.String(),
		NewValues: map[string]any{"end_date": endDate},
	})
	return &sg, nil
}

// ListStudentGroups returns memberships for a student (optionally current-only).
func ListStudentGroups(ctx context.Context, tx pgx.Tx, schoolID, studentID uuid.UUID, currentOnly bool) ([]model.StudentGroup, error) {
	rows, err := tx.Query(ctx, `
		SELECT sg.id, sg.school_id, sg.student_id, sg.group_id,
		       g.name,
		       sg.start_date, sg.end_date, sg.is_primary, sg.created_at
		FROM   student_groups sg
		JOIN   groups g ON g.id = sg.group_id
		WHERE  sg.school_id  = $1
		  AND  sg.student_id = $2
		  AND ($3 = FALSE OR sg.end_date IS NULL)
		ORDER BY sg.start_date DESC
	`, schoolID, studentID, currentOnly)
	if err != nil {
		return nil, fmt.Errorf("store.ListStudentGroups: %w", err)
	}
	defer rows.Close()

	var list []model.StudentGroup
	for rows.Next() {
		var sg model.StudentGroup
		if err := rows.Scan(
			&sg.ID, &sg.SchoolID, &sg.StudentID, &sg.GroupID,
			&sg.GroupName,
			&sg.StartDate, &sg.EndDate, &sg.IsPrimary, &sg.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("store.ListStudentGroups scan: %w", err)
		}
		sg.IsActive = sg.EndDate == nil
		list = append(list, sg)
	}
	return list, rows.Err()
}

// TeacherCanAccessStudent returns true if the teacher (by user_id) is assigned
// to at least one group that the student currently belongs to.
func TeacherCanAccessStudent(ctx context.Context, tx pgx.Tx, teacherUserID, studentID uuid.UUID) (bool, error) {
	var ok bool
	err := tx.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1
			FROM   student_groups sg
			JOIN   groups         g  ON g.id  = sg.group_id
			JOIN   teachers       t  ON t.id  = g.teacher_id
			WHERE  t.user_id    = $1
			  AND  sg.student_id = $2
			  AND  sg.end_date IS NULL
		)
	`, teacherUserID, studentID).Scan(&ok)
	if err != nil {
		return false, fmt.Errorf("store.TeacherCanAccessStudent: %w", err)
	}
	return ok, nil
}





