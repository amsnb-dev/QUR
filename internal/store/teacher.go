package store

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/quran-school/api/internal/model"
	"golang.org/x/crypto/bcrypt"
)

// CreateTeacher creates a user + teacher record in one transaction.
func CreateTeacher(ctx context.Context, tx pgx.Tx, schoolID uuid.UUID, roleID int, r *model.CreateTeacherRequest) (*model.Teacher, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(r.Password), 12)
	if err != nil {
		return nil, fmt.Errorf("store.CreateTeacher bcrypt: %w", err)
	}

	var userID uuid.UUID
	err = tx.QueryRow(ctx, `
		INSERT INTO users (school_id, role_id, full_name, email, password_hash, is_active)
		VALUES ($1, $2, $3, $4, $5, TRUE)
		RETURNING id
	`, schoolID, roleID, r.FullName, r.Email, string(hash)).Scan(&userID)
	if err != nil {
		return nil, fmt.Errorf("store.CreateTeacher insert user: %w", err)
	}

	salary := 0.0
	if r.BaseSalary != nil {
		salary = *r.BaseSalary
	}

	var t model.Teacher
	err = tx.QueryRow(ctx, `
		INSERT INTO teachers (school_id, user_id, specialization, qualification, hire_date, base_salary)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, school_id, user_id, specialization, qualification, hire_date,
		          base_salary, is_active, is_archived, archived_at, created_at, updated_at
	`, schoolID, userID, r.Specialization, r.Qualification, r.HireDate, salary,
	).Scan(
		&t.ID, &t.SchoolID, &t.UserID, &t.Specialization, &t.Qualification, &t.HireDate,
		&t.BaseSalary, &t.IsActive, &t.IsArchived, &t.ArchivedAt, &t.CreatedAt, &t.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("store.CreateTeacher insert teacher: %w", err)
	}
	t.FullName = r.FullName
	t.Email = r.Email
	return &t, nil
}

// ListTeachers returns teachers for the current tenant.
func ListTeachers(ctx context.Context, tx pgx.Tx, f model.ListTeachersFilter, p Page) ([]model.Teacher, int, error) {
	var total int
	if err := tx.QueryRow(ctx, `
		SELECT COUNT(*) FROM teachers WHERE ($1 OR is_archived = FALSE)
	`, f.IncludeArchived).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("store.ListTeachers count: %w", err)
	}

	rows, err := tx.Query(ctx, `
		SELECT t.id, t.school_id, t.user_id,
		       u.full_name, u.email,
		       t.specialization, t.qualification, t.hire_date,
		       t.base_salary, t.is_active, t.is_archived, t.archived_at,
		       t.created_at, t.updated_at
		FROM   teachers t
		JOIN   users u ON u.id = t.user_id
		WHERE  ($1 OR t.is_archived = FALSE)
		ORDER  BY u.full_name
		LIMIT  $2 OFFSET $3
	`, f.IncludeArchived, p.Limit, p.Offset)
	if err != nil {
		return nil, 0, fmt.Errorf("store.ListTeachers: %w", err)
	}
	defer rows.Close()

	var list []model.Teacher
	for rows.Next() {
		var t model.Teacher
		if err := rows.Scan(
			&t.ID, &t.SchoolID, &t.UserID,
			&t.FullName, &t.Email,
			&t.Specialization, &t.Qualification, &t.HireDate,
			&t.BaseSalary, &t.IsActive, &t.IsArchived, &t.ArchivedAt,
			&t.CreatedAt, &t.UpdatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("store.ListTeachers scan: %w", err)
		}
		list = append(list, t)
	}
	return list, total, rows.Err()
}

// GetTeacher fetches a single teacher by id.
func GetTeacher(ctx context.Context, tx pgx.Tx, id uuid.UUID) (*model.Teacher, error) {
	var t model.Teacher
	err := tx.QueryRow(ctx, `
		SELECT t.id, t.school_id, t.user_id,
		       u.full_name, u.email,
		       t.specialization, t.qualification, t.hire_date,
		       t.base_salary, t.is_active, t.is_archived, t.archived_at,
		       t.created_at, t.updated_at
		FROM   teachers t
		JOIN   users u ON u.id = t.user_id
		WHERE  t.id = $1
	`, id).Scan(
		&t.ID, &t.SchoolID, &t.UserID,
		&t.FullName, &t.Email,
		&t.Specialization, &t.Qualification, &t.HireDate,
		&t.BaseSalary, &t.IsActive, &t.IsArchived, &t.ArchivedAt,
		&t.CreatedAt, &t.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("store.GetTeacher: %w", err)
	}
	return &t, nil
}

// UpdateTeacher applies non-nil fields.
func UpdateTeacher(ctx context.Context, tx pgx.Tx, id uuid.UUID, r *model.UpdateTeacherRequest) (*model.Teacher, error) {
	cur, err := GetTeacher(ctx, tx, id)
	if err != nil || cur == nil {
		return cur, err
	}

	spec := cur.Specialization
	if r.Specialization != nil {
		spec = r.Specialization
	}
	qual := cur.Qualification
	if r.Qualification != nil {
		qual = r.Qualification
	}
	hire := cur.HireDate
	if r.HireDate != nil {
		hire = r.HireDate
	}
	salary := cur.BaseSalary
	if r.BaseSalary != nil {
		salary = *r.BaseSalary
	}
	isActive := cur.IsActive
	if r.IsActive != nil {
		isActive = *r.IsActive
	}

	if _, err = tx.Exec(ctx, `
		UPDATE teachers
		SET specialization = $2, qualification = $3, hire_date = $4,
		    base_salary = $5, is_active = $6
		WHERE id = $1
	`, id, spec, qual, hire, salary, isActive); err != nil {
		return nil, fmt.Errorf("store.UpdateTeacher: %w", err)
	}

	if r.FullName != nil {
		if _, err = tx.Exec(ctx, `UPDATE users SET full_name = $2 WHERE id = $1`, cur.UserID, *r.FullName); err != nil {
			return nil, fmt.Errorf("store.UpdateTeacher user: %w", err)
		}
	}

	return GetTeacher(ctx, tx, id)
}

// ArchiveTeacher soft-deletes a teacher.
func ArchiveTeacher(ctx context.Context, tx pgx.Tx, id uuid.UUID) error {
	tag, err := tx.Exec(ctx, `
		UPDATE teachers
		SET is_archived = TRUE, archived_at = NOW(), is_active = FALSE
		WHERE id = $1 AND is_archived = FALSE
	`, id)
	if err != nil {
		return fmt.Errorf("store.ArchiveTeacher: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// GetTeacherRoleID returns the role_id for "teacher".
func GetTeacherRoleID(ctx context.Context, tx pgx.Tx) (int, error) {
	var id int
	if err := tx.QueryRow(ctx, `SELECT id FROM roles WHERE name = 'teacher'`).Scan(&id); err != nil {
		return 0, fmt.Errorf("store.GetTeacherRoleID: %w", err)
	}
	return id, nil
}
