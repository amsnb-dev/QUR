package store

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/quran-school/api/internal/model"
	"golang.org/x/crypto/bcrypt"
)

func ListUsers(ctx context.Context, tx pgx.Tx, f model.ListUsersFilter, p Page) ([]model.UserRecord, int, error) {
	where := "WHERE 1=1"
	args := []any{}
	i := 1

	if !f.IncludeArchived {
		where += " AND u.is_active = TRUE"
	}
	if f.RoleName != "" {
		where += fmt.Sprintf(" AND r.name = $%d", i)
		args = append(args, f.RoleName)
		i++
	}

	var total int
	err := tx.QueryRow(ctx,
		`SELECT COUNT(*) FROM users u JOIN roles r ON r.id = u.role_id `+where,
		args...,
	).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	args = append(args, p.Limit, p.Offset)
	rows, err := tx.Query(ctx, `
		SELECT u.id, u.school_id, u.role_id, r.name, u.full_name, u.email, u.is_active, u.created_at
		FROM users u
		JOIN roles r ON r.id = u.role_id
		`+where+`
		ORDER BY u.created_at DESC
		LIMIT $`+fmt.Sprintf("%d", i)+` OFFSET $`+fmt.Sprintf("%d", i+1),
		args...,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var users []model.UserRecord
	for rows.Next() {
		var u model.UserRecord
		if err := rows.Scan(&u.ID, &u.SchoolID, &u.RoleID, &u.RoleName, &u.FullName, &u.Email, &u.IsActive, &u.CreatedAt); err != nil {
			return nil, 0, err
		}
		users = append(users, u)
	}
	return users, total, rows.Err()
}

func GetUser(ctx context.Context, tx pgx.Tx, id uuid.UUID) (*model.UserRecord, error) {
	var u model.UserRecord
	err := tx.QueryRow(ctx, `
		SELECT u.id, u.school_id, u.role_id, r.name, u.full_name, u.email, u.is_active, u.created_at
		FROM users u JOIN roles r ON r.id = u.role_id
		WHERE u.id = $1`, id,
	).Scan(&u.ID, &u.SchoolID, &u.RoleID, &u.RoleName, &u.FullName, &u.Email, &u.IsActive, &u.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func CreateUser(ctx context.Context, tx pgx.Tx, schoolID *uuid.UUID, req *model.CreateUserRequest) (*model.UserRecord, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), 12)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	var roleID int16
	err = tx.QueryRow(ctx, `SELECT id FROM roles WHERE name = $1`, req.Role).Scan(&roleID)
	if err != nil {
		return nil, fmt.Errorf("role %q not found", req.Role)
	}

	id := uuid.New()
	_, err = tx.Exec(ctx, `
		INSERT INTO users (id, school_id, role_id, full_name, email, password_hash)
		VALUES ($1, $2, $3, $4, $5, $6)`,
		id, schoolID, roleID, req.FullName, req.Email, string(hash),
	)
	if err != nil {
		return nil, fmt.Errorf("insert user: %w", err)
	}

	return GetUser(ctx, tx, id)
}

func UpdateUser(ctx context.Context, tx pgx.Tx, id uuid.UUID, req *model.UpdateUserRequest) (*model.UserRecord, error) {
	if req.FullName != nil {
		if _, err := tx.Exec(ctx, `UPDATE users SET full_name=$1 WHERE id=$2`, *req.FullName, id); err != nil {
			return nil, err
		}
	}
	if req.Email != nil {
		if _, err := tx.Exec(ctx, `UPDATE users SET email=$1 WHERE id=$2`, *req.Email, id); err != nil {
			return nil, err
		}
	}
	if req.IsActive != nil {
		if _, err := tx.Exec(ctx, `UPDATE users SET is_active=$1 WHERE id=$2`, *req.IsActive, id); err != nil {
			return nil, err
		}
	}
	if req.Password != nil {
		hash, err := bcrypt.GenerateFromPassword([]byte(*req.Password), 12)
		if err != nil {
			return nil, err
		}
		if _, err := tx.Exec(ctx, `UPDATE users SET password_hash=$1 WHERE id=$2`, string(hash), id); err != nil {
			return nil, err
		}
	}
	return GetUser(ctx, tx, id)
}

func ArchiveUser(ctx context.Context, tx pgx.Tx, id uuid.UUID) error {
	tag, err := tx.Exec(ctx, `UPDATE users SET is_active=FALSE, is_archived=TRUE, archived_at=NOW() WHERE id=$1 AND is_active=TRUE`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}
