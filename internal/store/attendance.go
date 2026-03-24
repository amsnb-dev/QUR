package store

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/quran-school/api/internal/model"
)

// UpsertAttendanceItem upserts one attendance row within an existing tx.
// ON CONFLICT uses the unique constraint (school_id, student_id, group_id, date).
func UpsertAttendanceItem(ctx context.Context, tx pgx.Tx, schoolID, groupID, recordedBy uuid.UUID, item model.AttendanceItem, date time.Time) (*model.AttendanceRecord, error) {
	var r model.AttendanceRecord
	err := tx.QueryRow(ctx, `
		INSERT INTO attendance
		       (school_id, student_id, group_id, recorded_by, date, status, note)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (school_id, student_id, group_id, date)
		DO UPDATE SET
		       status      = EXCLUDED.status,
		       note        = EXCLUDED.note,
		       recorded_by = EXCLUDED.recorded_by,
		       updated_at  = NOW()
		RETURNING id, school_id, student_id, group_id, recorded_by,
		          date, status, note, created_at, updated_at
	`, schoolID, item.StudentID, groupID, recordedBy, date, item.Status, item.Note,
	).Scan(
		&r.ID, &r.SchoolID, &r.StudentID, &r.GroupID, &r.RecordedBy,
		&r.Date, &r.Status, &r.Note, &r.CreatedAt, &r.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("store.UpsertAttendanceItem: %w", err)
	}
	return &r, nil
}

// ListAttendance returns paginated attendance records for a group on a date.
func ListAttendance(ctx context.Context, tx pgx.Tx, groupID uuid.UUID, date *time.Time, p Page) ([]model.AttendanceRecord, int, error) {
	var total int
	if err := tx.QueryRow(ctx, `
		SELECT COUNT(*) FROM attendance
		WHERE  group_id = $1
		  AND ($2::DATE IS NULL OR date = $2)
	`, groupID, date).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("store.ListAttendance count: %w", err)
	}

	rows, err := tx.Query(ctx, `
		SELECT id, school_id, student_id, group_id, recorded_by,
		       date, status, note, created_at, updated_at
		FROM   attendance
		WHERE  group_id = $1
		  AND ($2::DATE IS NULL OR date = $2)
		ORDER BY date DESC, student_id
		LIMIT  $3 OFFSET $4
	`, groupID, date, p.Limit, p.Offset)
	if err != nil {
		return nil, 0, fmt.Errorf("store.ListAttendance: %w", err)
	}
	defer rows.Close()

	var list []model.AttendanceRecord
	for rows.Next() {
		var r model.AttendanceRecord
		if err := rows.Scan(
			&r.ID, &r.SchoolID, &r.StudentID, &r.GroupID, &r.RecordedBy,
			&r.Date, &r.Status, &r.Note, &r.CreatedAt, &r.UpdatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("store.ListAttendance scan: %w", err)
		}
		list = append(list, r)
	}
	return list, total, rows.Err()
}

// UpdateAttendance updates status/note on an existing attendance record.
func UpdateAttendance(ctx context.Context, tx pgx.Tx, id uuid.UUID, r *model.UpdateAttendanceRequest) (*model.AttendanceRecord, error) {
	var rec model.AttendanceRecord
	err := tx.QueryRow(ctx, `
		UPDATE attendance
		SET    status = COALESCE($2, status),
		       note   = COALESCE($3, note)
		WHERE  id = $1
		RETURNING id, school_id, student_id, group_id, recorded_by,
		          date, status, note, created_at, updated_at
	`, id, r.Status, r.Note).Scan(
		&rec.ID, &rec.SchoolID, &rec.StudentID, &rec.GroupID, &rec.RecordedBy,
		&rec.Date, &rec.Status, &rec.Note, &rec.CreatedAt, &rec.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("store.UpdateAttendance: %w", err)
	}
	return &rec, nil
}

// GetAttendanceSchoolID returns the school_id of an attendance record (for RBAC).
func GetAttendanceSchoolID(ctx context.Context, tx pgx.Tx, id uuid.UUID) (uuid.UUID, error) {
	var sid uuid.UUID
	err := tx.QueryRow(ctx, `SELECT school_id FROM attendance WHERE id = $1`, id).Scan(&sid)
	if err == pgx.ErrNoRows {
		return uuid.Nil, ErrNotFound
	}
	return sid, err
}
