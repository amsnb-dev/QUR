package store

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/quran-school/api/internal/model"
)

// CreateMemorization inserts a new memorization_logs row.
func CreateMemorization(ctx context.Context, tx pgx.Tx, schoolID, studentID, recordedBy uuid.UUID, r *model.CreateMemorizationRequest) (*model.MemorizationRecord, error) {
	var rec model.MemorizationRecord
	err := tx.QueryRow(ctx, `
		INSERT INTO memorization_logs
		       (school_id, student_id, group_id, recorded_by,
		        date, surah_number, from_verse, to_verse,
		        entry_type, grade, notes)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
		RETURNING id, school_id, student_id, group_id, recorded_by,
		          date, surah_number, from_verse, to_verse,
		          entry_type, grade, notes, created_at
	`, schoolID, studentID, r.GroupID, recordedBy,
		r.Date, r.SurahNumber, r.FromVerse, r.ToVerse,
		r.EntryType, r.Grade, r.Notes,
	).Scan(
		&rec.ID, &rec.SchoolID, &rec.StudentID, &rec.GroupID, &rec.RecordedBy,
		&rec.Date, &rec.SurahNumber, &rec.FromVerse, &rec.ToVerse,
		&rec.EntryType, &rec.Grade, &rec.Notes, &rec.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("store.CreateMemorization: %w", err)
	}
	return &rec, nil
}

// ListMemorization returns paginated memorization records for a student.
func ListMemorization(ctx context.Context, tx pgx.Tx, studentID uuid.UUID, p Page) ([]model.MemorizationRecord, int, error) {
	var total int
	if err := tx.QueryRow(ctx, `
		SELECT COUNT(*) FROM memorization_logs WHERE student_id = $1
	`, studentID).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("store.ListMemorization count: %w", err)
	}

	rows, err := tx.Query(ctx, `
		SELECT id, school_id, student_id, group_id, recorded_by,
		       date, surah_number, from_verse, to_verse,
		       entry_type, grade, notes, created_at
		FROM   memorization_logs
		WHERE  student_id = $1
		ORDER BY date DESC, created_at DESC
		LIMIT  $2 OFFSET $3
	`, studentID, p.Limit, p.Offset)
	if err != nil {
		return nil, 0, fmt.Errorf("store.ListMemorization: %w", err)
	}
	defer rows.Close()

	var list []model.MemorizationRecord
	for rows.Next() {
		var rec model.MemorizationRecord
		if err := rows.Scan(
			&rec.ID, &rec.SchoolID, &rec.StudentID, &rec.GroupID, &rec.RecordedBy,
			&rec.Date, &rec.SurahNumber, &rec.FromVerse, &rec.ToVerse,
			&rec.EntryType, &rec.Grade, &rec.Notes, &rec.CreatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("store.ListMemorization scan: %w", err)
		}
		list = append(list, rec)
	}
	return list, total, rows.Err()
}

// UpdateMemorization patches grade / notes / verse range.
func UpdateMemorization(ctx context.Context, tx pgx.Tx, id uuid.UUID, r *model.UpdateMemorizationRequest) (*model.MemorizationRecord, error) {
	var rec model.MemorizationRecord
	err := tx.QueryRow(ctx, `
		UPDATE memorization_logs
		SET    grade      = COALESCE($2, grade),
		       notes      = COALESCE($3, notes),
		       from_verse = COALESCE($4, from_verse),
		       to_verse   = COALESCE($5, to_verse)
		WHERE  id = $1
		RETURNING id, school_id, student_id, group_id, recorded_by,
		          date, surah_number, from_verse, to_verse,
		          entry_type, grade, notes, created_at
	`, id, r.Grade, r.Notes, r.FromVerse, r.ToVerse,
	).Scan(
		&rec.ID, &rec.SchoolID, &rec.StudentID, &rec.GroupID, &rec.RecordedBy,
		&rec.Date, &rec.SurahNumber, &rec.FromVerse, &rec.ToVerse,
		&rec.EntryType, &rec.Grade, &rec.Notes, &rec.CreatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("store.UpdateMemorization: %w", err)
	}
	return &rec, nil
}

// GetMemorizationSchoolAndStudent returns school_id + student_id for RBAC checks.
func GetMemorizationSchoolAndStudent(ctx context.Context, tx pgx.Tx, id uuid.UUID) (schoolID, studentID uuid.UUID, err error) {
	err = tx.QueryRow(ctx,
		`SELECT school_id, student_id FROM memorization_logs WHERE id = $1`, id,
	).Scan(&schoolID, &studentID)
	if err == pgx.ErrNoRows {
		return uuid.Nil, uuid.Nil, ErrNotFound
	}
	return schoolID, studentID, err
}
