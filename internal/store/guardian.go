package store

import (
"context"
"fmt"

"github.com/google/uuid"
"github.com/jackc/pgx/v5"
"github.com/quran-school/api/internal/model"
)

func CreateGuardian(ctx context.Context, tx pgx.Tx, schoolID uuid.UUID, r *model.CreateGuardianRequest) (*model.Guardian, error) {
var g model.Guardian
err := tx.QueryRow(ctx, `
INSERT INTO guardians (school_id, full_name, phone, phone2, email, address, national_id, username, password_plain, notes)
VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
RETURNING id, school_id, full_name, phone, phone2, email, address, national_id, username, password_plain, notes, relation, is_active, created_at, updated_at
`, schoolID, r.FullName, r.Phone, r.Phone2, r.Email, r.Address, r.NationalID, r.Username, r.PasswordPlain, r.Notes,
).Scan(&g.ID, &g.SchoolID, &g.FullName, &g.Phone, &g.Phone2, &g.Email, &g.Address, &g.NationalID, &g.Username, &g.PasswordPlain, &g.Notes, &g.Relation, &g.IsActive, &g.CreatedAt, &g.UpdatedAt)
if err != nil {
return nil, fmt.Errorf("store.CreateGuardian: %w", err)
}
return &g, nil
}

func GetGuardian(ctx context.Context, tx pgx.Tx, id uuid.UUID) (*model.Guardian, error) {
var g model.Guardian
err := tx.QueryRow(ctx, `
SELECT id, school_id, full_name, phone, phone2, email, address, national_id, username, password_plain, notes, relation, is_active, created_at, updated_at
FROM guardians WHERE id = $1
`, id).Scan(&g.ID, &g.SchoolID, &g.FullName, &g.Phone, &g.Phone2, &g.Email, &g.Address, &g.NationalID, &g.Username, &g.PasswordPlain, &g.Notes, &g.Relation, &g.IsActive, &g.CreatedAt, &g.UpdatedAt)
if err != nil {
return nil, fmt.Errorf("store.GetGuardian: %w", err)
}
return &g, nil
}

func ListGuardians(ctx context.Context, tx pgx.Tx, f model.ListGuardiansFilter) ([]model.Guardian, error) {
rows, err := tx.Query(ctx, `
SELECT id, school_id, full_name, phone, phone2, email, address, national_id, username, password_plain, notes, relation, is_active, created_at, updated_at
FROM guardians
WHERE ($1 = '' OR full_name ILIKE '%' || $1 || '%' OR phone ILIKE '%' || $1 || '%')
ORDER BY full_name
`, f.Search)
if err != nil {
return nil, fmt.Errorf("store.ListGuardians: %w", err)
}
defer rows.Close()
var list []model.Guardian
for rows.Next() {
var g model.Guardian
if err := rows.Scan(&g.ID, &g.SchoolID, &g.FullName, &g.Phone, &g.Phone2, &g.Email, &g.Address, &g.NationalID, &g.Username, &g.PasswordPlain, &g.Notes, &g.Relation, &g.IsActive, &g.CreatedAt, &g.UpdatedAt); err != nil {
return nil, err
}
list = append(list, g)
}
return list, nil
}

func UpdateGuardian(ctx context.Context, tx pgx.Tx, id uuid.UUID, r *model.UpdateGuardianRequest) (*model.Guardian, error) {
cur, err := GetGuardian(ctx, tx, id)
if err != nil {
return nil, err
}
fullName := cur.FullName
if r.FullName != nil { fullName = *r.FullName }
phone := cur.Phone
if r.Phone != nil { phone = r.Phone }
phone2 := cur.Phone2
if r.Phone2 != nil { phone2 = r.Phone2 }
email := cur.Email
if r.Email != nil { email = r.Email }
address := cur.Address
if r.Address != nil { address = r.Address }
nationalID := cur.NationalID
if r.NationalID != nil { nationalID = r.NationalID }
username := cur.Username
if r.Username != nil { username = r.Username }
passwordPlain := cur.PasswordPlain
if r.PasswordPlain != nil { passwordPlain = r.PasswordPlain }
notes := cur.Notes
if r.Notes != nil { notes = r.Notes }
isActive := cur.IsActive
if r.IsActive != nil { isActive = *r.IsActive }

var g model.Guardian
err = tx.QueryRow(ctx, `
UPDATE guardians SET full_name=$2, phone=$3, phone2=$4, email=$5, address=$6, national_id=$7, username=$8, password_plain=$9, notes=$10, is_active=$11
WHERE id=$1
RETURNING id, school_id, full_name, phone, phone2, email, address, national_id, username, password_plain, notes, relation, is_active, created_at, updated_at
`, id, fullName, phone, phone2, email, address, nationalID, username, passwordPlain, notes, isActive,
).Scan(&g.ID, &g.SchoolID, &g.FullName, &g.Phone, &g.Phone2, &g.Email, &g.Address, &g.NationalID, &g.Username, &g.PasswordPlain, &g.Notes, &g.Relation, &g.IsActive, &g.CreatedAt, &g.UpdatedAt)
if err != nil {
return nil, fmt.Errorf("store.UpdateGuardian: %w", err)
}
return &g, nil
}
