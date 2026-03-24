package store

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"
	"github.com/quran-school/api/internal/model"
)

func ListRoles(ctx context.Context, tx pgx.Tx) ([]model.SchoolRole, error) {
	q := `SELECT id,school_id,name,slug,permissions,is_system,created_at FROM school_roles ORDER BY name`
	rows, err := tx.Query(ctx, q)
	if err != nil { return nil, fmt.Errorf("store.ListRoles: %w", err) }
	defer rows.Close()
	var list []model.SchoolRole
	for rows.Next() {
		var r model.SchoolRole
		if err := rows.Scan(&r.ID,&r.SchoolID,&r.Name,&r.Slug,&r.Permissions,&r.IsSystem,&r.CreatedAt); err != nil { return nil, err }
		list = append(list, r)
	}
	return list, nil
}

func UpdateRolePermissions(ctx context.Context, tx pgx.Tx, id uuid.UUID, perms json.RawMessage) (*model.SchoolRole, error) {
	var r model.SchoolRole
	q := `UPDATE school_roles SET permissions=$2 WHERE id=$1 RETURNING id,school_id,name,slug,permissions,is_system,created_at`
	err := tx.QueryRow(ctx, q, id, perms).Scan(&r.ID,&r.SchoolID,&r.Name,&r.Slug,&r.Permissions,&r.IsSystem,&r.CreatedAt)
	if err != nil { return nil, fmt.Errorf("store.UpdateRolePermissions: %w", err) }
	return &r, nil
}

func ListStaff(ctx context.Context, tx pgx.Tx) ([]model.SchoolStaff, error) {
	q := `SELECT s.id,s.school_id,s.role_id,r.name,r.slug,s.full_name,s.username,s.password_plain,s.is_active,s.created_at,s.updated_at FROM school_staff s JOIN school_roles r ON r.id=s.role_id ORDER BY s.full_name`
	rows, err := tx.Query(ctx, q)
	if err != nil { return nil, fmt.Errorf("store.ListStaff: %w", err) }
	defer rows.Close()
	var list []model.SchoolStaff
	for rows.Next() {
		var s model.SchoolStaff
		if err := rows.Scan(&s.ID,&s.SchoolID,&s.RoleID,&s.RoleName,&s.RoleSlug,&s.FullName,&s.Username,&s.PasswordPlain,&s.IsActive,&s.CreatedAt,&s.UpdatedAt); err != nil { return nil, err }
		list = append(list, s)
	}
	return list, nil
}

func CreateStaff(ctx context.Context, tx pgx.Tx, schoolID uuid.UUID, r *model.CreateStaffRequest) (*model.SchoolStaff, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(r.Password), bcrypt.DefaultCost)
	if err != nil { return nil, err }
	roleID, err := uuid.Parse(r.RoleID)
	if err != nil { return nil, fmt.Errorf("invalid role_id: %w", err) }
	var s model.SchoolStaff
	q := `INSERT INTO school_staff (school_id,role_id,full_name,username,password_hash,password_plain,is_active) VALUES ($1,$2,$3,$4,$5,$6,true) RETURNING id,school_id,role_id,full_name,username,password_plain,is_active,created_at,updated_at`
	err = tx.QueryRow(ctx, q, schoolID, roleID, r.FullName, r.Username, string(hash), r.PasswordPlain).Scan(&s.ID,&s.SchoolID,&s.RoleID,&s.FullName,&s.Username,&s.PasswordPlain,&s.IsActive,&s.CreatedAt,&s.UpdatedAt)
	if err != nil { return nil, fmt.Errorf("store.CreateStaff: %w", err) }
	return &s, nil
}

func UpdateStaff(ctx context.Context, tx pgx.Tx, id uuid.UUID, r *model.UpdateStaffRequest) (*model.SchoolStaff, error) {
	sets := "updated_at=now()"
	args := []any{id}
	n := 2
	if r.FullName != nil      { sets += fmt.Sprintf(",full_name=$%d", n);      args = append(args, *r.FullName);      n++ }
	if r.Username != nil      { sets += fmt.Sprintf(",username=$%d", n);        args = append(args, *r.Username);      n++ }
	if r.IsActive != nil      { sets += fmt.Sprintf(",is_active=$%d", n);       args = append(args, *r.IsActive);      n++ }
	if r.PasswordPlain != nil { sets += fmt.Sprintf(",password_plain=$%d", n);  args = append(args, *r.PasswordPlain); n++ }
	if r.Password != nil {
		hash, _ := bcrypt.GenerateFromPassword([]byte(*r.Password), bcrypt.DefaultCost)
		sets += fmt.Sprintf(",password_hash=$%d", n); args = append(args, string(hash)); n++
	}
	if r.RoleID != nil {
		rid, _ := uuid.Parse(*r.RoleID)
		sets += fmt.Sprintf(",role_id=$%d", n); args = append(args, rid); n++
	}
	var s model.SchoolStaff
	q := fmt.Sprintf("UPDATE school_staff SET %s WHERE id=$1 RETURNING id,school_id,role_id,full_name,username,password_plain,is_active,created_at,updated_at", sets)
	err := tx.QueryRow(ctx, q, args...).Scan(&s.ID,&s.SchoolID,&s.RoleID,&s.FullName,&s.Username,&s.PasswordPlain,&s.IsActive,&s.CreatedAt,&s.UpdatedAt)
	if err != nil { return nil, fmt.Errorf("store.UpdateStaff: %w", err) }
	return &s, nil
}

func GetSettings(ctx context.Context, tx pgx.Tx, schoolID uuid.UUID) (*model.SchoolSettings, error) {
	var s model.SchoolSettings
	q := `SELECT school_id,currency,language,timezone,date_format,logo_url,address,phone,email,updated_at FROM school_settings WHERE school_id=$1`
	err := tx.QueryRow(ctx, q, schoolID).Scan(&s.SchoolID,&s.Currency,&s.Language,&s.Timezone,&s.DateFormat,&s.LogoURL,&s.Address,&s.Phone,&s.Email,&s.UpdatedAt)
	if err == pgx.ErrNoRows { return &model.SchoolSettings{SchoolID:schoolID,Currency:"SAR",Language:"ar",Timezone:"Asia/Riyadh",DateFormat:"hijri"}, nil }
	if err != nil { return nil, fmt.Errorf("store.GetSettings: %w", err) }
	return &s, nil
}

func UpsertSettings(ctx context.Context, tx pgx.Tx, schoolID uuid.UUID, r *model.UpdateSettingsRequest) (*model.SchoolSettings, error) {
	var s model.SchoolSettings
	q := `INSERT INTO school_settings (school_id,currency,language,timezone,date_format,logo_url,address,phone,email) VALUES ($1,COALESCE($2,'SAR'),COALESCE($3,'ar'),COALESCE($4,'Asia/Riyadh'),COALESCE($5,'hijri'),$6,$7,$8,$9) ON CONFLICT (school_id) DO UPDATE SET currency=COALESCE(EXCLUDED.currency,school_settings.currency),language=COALESCE(EXCLUDED.language,school_settings.language),timezone=COALESCE(EXCLUDED.timezone,school_settings.timezone),date_format=COALESCE(EXCLUDED.date_format,school_settings.date_format),logo_url=COALESCE(EXCLUDED.logo_url,school_settings.logo_url),address=COALESCE(EXCLUDED.address,school_settings.address),phone=COALESCE(EXCLUDED.phone,school_settings.phone),email=COALESCE(EXCLUDED.email,school_settings.email),updated_at=now() RETURNING school_id,currency,language,timezone,date_format,logo_url,address,phone,email,updated_at`
	err := tx.QueryRow(ctx, q, schoolID,r.Currency,r.Language,r.Timezone,r.DateFormat,r.LogoURL,r.Address,r.Phone,r.Email).Scan(&s.SchoolID,&s.Currency,&s.Language,&s.Timezone,&s.DateFormat,&s.LogoURL,&s.Address,&s.Phone,&s.Email,&s.UpdatedAt)
	if err != nil { return nil, fmt.Errorf("store.UpsertSettings: %w", err) }
	return &s, nil
}