package store

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/quran-school/api/internal/model"
)

const schoolCols = `
	s.id, s.name, s.city, s.country, s.plan, s.is_active,
	s.default_monthly_fee, s.created_at, s.updated_at,
	COUNT(DISTINCT st.id) FILTER (WHERE st.is_archived = FALSE),
	COUNT(DISTINCT t.id)  FILTER (WHERE t.is_archived  = FALSE),
	COUNT(DISTINCT g.id)  FILTER (WHERE g.is_archived  = FALSE)
`
const schoolJoins = `
	FROM schools s
	LEFT JOIN students st ON st.school_id = s.id
	LEFT JOIN teachers  t ON t.school_id  = s.id
	LEFT JOIN groups    g ON g.school_id  = s.id
`

func scanSchoolRow(row pgx.Row) (*model.School, error) {
	var sc model.School
	err := row.Scan(
		&sc.ID, &sc.Name, &sc.City, &sc.Country, &sc.Plan, &sc.IsActive,
		&sc.DefaultMonthlyFee, &sc.CreatedAt, &sc.UpdatedAt,
		&sc.StudentCount, &sc.TeacherCount, &sc.GroupCount,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &sc, nil
}

// ListSchools — uses the tenant tx (is_super_admin=1 bypasses RLS on joined tables).
func ListSchools(ctx context.Context, tx pgx.Tx) ([]model.School, error) {
	rows, err := tx.Query(ctx,
		`SELECT `+schoolCols+schoolJoins+` GROUP BY s.id ORDER BY s.created_at DESC`)
	if err != nil {
		return nil, fmt.Errorf("store.ListSchools: %w", err)
	}
	defer rows.Close()

	var list []model.School
	for rows.Next() {
		sc, err := scanSchoolRow(rows)
		if err != nil {
			return nil, fmt.Errorf("store.ListSchools scan: %w", err)
		}
		list = append(list, *sc)
	}
	return list, rows.Err()
}

// GetSchool — uses the tenant tx.
func GetSchool(ctx context.Context, tx pgx.Tx, id uuid.UUID) (*model.School, error) {
	sc, err := scanSchoolRow(tx.QueryRow(ctx,
		`SELECT `+schoolCols+schoolJoins+` WHERE s.id = $1 GROUP BY s.id`, id))
	if err != nil {
		return nil, fmt.Errorf("store.GetSchool: %w", err)
	}
	return sc, nil
}

// CreateSchool — super_admin tx.
func CreateSchool(ctx context.Context, tx pgx.Tx, r *model.CreateSchoolRequest) (*model.School, error) {
	country := "DZ"
	if r.Country != nil {
		country = *r.Country
	}
	plan := "trial"
	if r.Plan != nil {
		plan = *r.Plan
	}
	fee := 0.0
	if r.DefaultMonthlyFee != nil {
		fee = *r.DefaultMonthlyFee
	}

	var sc model.School
	err := tx.QueryRow(ctx, `
		INSERT INTO schools (name, city, country, plan, default_monthly_fee)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, name, city, country, plan, is_active,
		          default_monthly_fee, created_at, updated_at
	`, r.Name, r.City, country, plan, fee).Scan(
		&sc.ID, &sc.Name, &sc.City, &sc.Country, &sc.Plan, &sc.IsActive,
		&sc.DefaultMonthlyFee, &sc.CreatedAt, &sc.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("store.CreateSchool: %w", err)
	}
	return &sc, nil
}

// UpdateSchool — super_admin tx.
func UpdateSchool(ctx context.Context, tx pgx.Tx, id uuid.UUID, r *model.UpdateSchoolRequest) (*model.School, error) {
	var cur model.School
	err := tx.QueryRow(ctx, `
		SELECT id, name, city, country, plan, is_active, default_monthly_fee, created_at, updated_at
		FROM schools WHERE id = $1
	`, id).Scan(
		&cur.ID, &cur.Name, &cur.City, &cur.Country, &cur.Plan, &cur.IsActive,
		&cur.DefaultMonthlyFee, &cur.CreatedAt, &cur.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("store.UpdateSchool fetch: %w", err)
	}

	name := cur.Name
	if r.Name != nil { name = *r.Name }
	city := cur.City
	if r.City != nil { city = r.City }
	country := cur.Country
	if r.Country != nil { country = *r.Country }
	plan := cur.Plan
	if r.Plan != nil { plan = *r.Plan }
	isActive := cur.IsActive
	if r.IsActive != nil { isActive = *r.IsActive }
	fee := cur.DefaultMonthlyFee
	if r.DefaultMonthlyFee != nil { fee = *r.DefaultMonthlyFee }

	var sc model.School
	err = tx.QueryRow(ctx, `
		UPDATE schools
		SET name = $2, city = $3, country = $4, plan = $5,
		    is_active = $6, default_monthly_fee = $7
		WHERE id = $1
		RETURNING id, name, city, country, plan, is_active,
		          default_monthly_fee, created_at, updated_at
	`, id, name, city, country, plan, isActive, fee).Scan(
		&sc.ID, &sc.Name, &sc.City, &sc.Country, &sc.Plan, &sc.IsActive,
		&sc.DefaultMonthlyFee, &sc.CreatedAt, &sc.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("store.UpdateSchool: %w", err)
	}
	return &sc, nil
}
