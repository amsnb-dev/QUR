package store

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/quran-school/api/internal/model"
)

// ── Groups ──────────────────────────────────────────────────────────────────
func CreateGroup(ctx context.Context, tx pgx.Tx, schoolID uuid.UUID, r *model.CreateGroupRequest) (*model.Group, error) {
	days := r.Days
	if days == nil {
		days = []int16{}
	}
	var g model.Group
	err := tx.QueryRow(ctx, `
		INSERT INTO groups
		       (school_id, teacher_id, name, level, stage, room, days, start_time, end_time)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, school_id, teacher_id, name, level, stage, room, days,
		          start_time::TEXT, end_time::TEXT,
		          is_active, is_archived, archived_at, created_at, updated_at
	`, schoolID, r.TeacherID, r.Name, r.Level, r.Stage, r.Room, days, r.StartTime, r.EndTime,
	).Scan(
		&g.ID, &g.SchoolID, &g.TeacherID, &g.Name, &g.Level, &g.Stage, &g.Room, &g.Days,
		&g.StartTime, &g.EndTime,
		&g.IsActive, &g.IsArchived, &g.ArchivedAt, &g.CreatedAt, &g.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("store.CreateGroup: %w", err)
	}
	return &g, nil
}

func ListGroups(ctx context.Context, tx pgx.Tx, f model.ListGroupsFilter, p Page) ([]model.Group, int, error) {
	var total int
	if err := tx.QueryRow(ctx, `
		SELECT COUNT(*) FROM groups WHERE ($1 OR is_archived = FALSE)
	`, f.IncludeArchived).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("store.ListGroups count: %w", err)
	}
	rows, err := tx.Query(ctx, `
		SELECT g.id, g.school_id, g.teacher_id, g.name, g.level, g.stage, g.room, g.days,
		       g.start_time::TEXT, g.end_time::TEXT,
		       g.is_active, g.is_archived, g.archived_at, g.created_at, g.updated_at,
		       u.full_name,
		       COUNT(sg.id) FILTER (WHERE sg.end_date IS NULL)
		FROM   groups g
		LEFT JOIN teachers      t  ON t.id  = g.teacher_id
		LEFT JOIN users         u  ON u.id  = t.user_id
		LEFT JOIN student_groups sg ON sg.group_id = g.id
		WHERE ($1 OR g.is_archived = FALSE)
		GROUP BY g.id, u.full_name
		ORDER BY g.name
		LIMIT  $2 OFFSET $3
	`, f.IncludeArchived, p.Limit, p.Offset)
	if err != nil {
		return nil, 0, fmt.Errorf("store.ListGroups: %w", err)
	}
	defer rows.Close()
	var list []model.Group
	for rows.Next() {
		var g model.Group
		if err := rows.Scan(
			&g.ID, &g.SchoolID, &g.TeacherID, &g.Name, &g.Level, &g.Stage, &g.Room, &g.Days,
			&g.StartTime, &g.EndTime,
			&g.IsActive, &g.IsArchived, &g.ArchivedAt, &g.CreatedAt, &g.UpdatedAt,
			&g.TeacherName, &g.StudentsCount,
		); err != nil {
			return nil, 0, fmt.Errorf("store.ListGroups scan: %w", err)
		}
		list = append(list, g)
	}
	return list, total, rows.Err()
}

func GetGroup(ctx context.Context, tx pgx.Tx, id uuid.UUID) (*model.Group, error) {
	var g model.Group
	err := tx.QueryRow(ctx, `
		SELECT g.id, g.school_id, g.teacher_id, g.name, g.level, g.stage, g.room, g.days,
		       g.start_time::TEXT, g.end_time::TEXT,
		       g.is_active, g.is_archived, g.archived_at, g.created_at, g.updated_at,
		       u.full_name,
		       COUNT(sg.id) FILTER (WHERE sg.end_date IS NULL)
		FROM   groups g
		LEFT JOIN teachers      t  ON t.id  = g.teacher_id
		LEFT JOIN users         u  ON u.id  = t.user_id
		LEFT JOIN student_groups sg ON sg.group_id = g.id
		WHERE  g.id = $1
		GROUP BY g.id, u.full_name
	`, id).Scan(
		&g.ID, &g.SchoolID, &g.TeacherID, &g.Name, &g.Level, &g.Stage, &g.Room, &g.Days,
		&g.StartTime, &g.EndTime,
		&g.IsActive, &g.IsArchived, &g.ArchivedAt, &g.CreatedAt, &g.UpdatedAt,
		&g.TeacherName, &g.StudentsCount,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("store.GetGroup: %w", err)
	}
	return &g, nil
}

func UpdateGroup(ctx context.Context, tx pgx.Tx, id uuid.UUID, r *model.UpdateGroupRequest) (*model.Group, error) {
	cur, err := GetGroup(ctx, tx, id)
	if err != nil || cur == nil {
		return cur, err
	}

	teacherID := cur.TeacherID
	if r.TeacherID != nil {
		teacherID = r.TeacherID
	}
	name := cur.Name
	if r.Name != nil {
		name = *r.Name
	}
	level := cur.Level
	if r.Level != nil {
		level = r.Level
	}
	stage := cur.Stage
	if r.Stage != nil {
		stage = r.Stage
	}
	room := cur.Room
	if r.Room != nil {
		room = r.Room
	}
	days := cur.Days
	if r.Days != nil {
		days = r.Days
	}
	startTime := cur.StartTime
	if r.StartTime != nil {
		startTime = r.StartTime
	}
	endTime := cur.EndTime
	if r.EndTime != nil {
		endTime = r.EndTime
	}
	isActive := cur.IsActive
	if r.IsActive != nil {
		isActive = *r.IsActive
	}

	var g model.Group
	err = tx.QueryRow(ctx, `
		UPDATE groups
		SET    teacher_id = $2,
		       name       = $3,
		       level      = $4,
		       stage      = $5,
		       room       = $6,
		       days       = $7,
		       start_time = $8,
		       end_time   = $9,
		       is_active  = $10
		WHERE  id = $1
		RETURNING id, school_id, teacher_id, name, level, stage, room, days,
		          start_time::TEXT, end_time::TEXT,
		          is_active, is_archived, archived_at, created_at, updated_at
	`, id, teacherID, name, level, stage, room, days, startTime, endTime, isActive,
	).Scan(
		&g.ID, &g.SchoolID, &g.TeacherID, &g.Name, &g.Level, &g.Stage, &g.Room, &g.Days,
		&g.StartTime, &g.EndTime,
		&g.IsActive, &g.IsArchived, &g.ArchivedAt, &g.CreatedAt, &g.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("store.UpdateGroup: %w", err)
	}
	return &g, nil
}

func ArchiveGroup(ctx context.Context, tx pgx.Tx, id uuid.UUID, userID uuid.UUID, schoolID *uuid.UUID) error {
	tag, err := tx.Exec(ctx, `
		UPDATE groups
		SET is_archived = TRUE,
		    archived_at = NOW(),
		    is_active   = FALSE
		WHERE id          = $1
		  AND is_archived = FALSE
	`, id)
	if err != nil {
		return fmt.Errorf("store.ArchiveGroup: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return InsertAudit(ctx, tx, AuditParams{
		SchoolID:  schoolID,
		UserID:    &userID,
		Action:    "ARCHIVE",
		TableName: "groups",
		RecordID:  id.String(),
		NewValues: map[string]any{"is_archived": true},
	})
}

func GetGroupTeacherUserID(ctx context.Context, tx pgx.Tx, groupID uuid.UUID) (uuid.UUID, error) {
	var userID uuid.UUID
	err := tx.QueryRow(ctx, `
		SELECT t.user_id
		FROM   groups g
		JOIN   teachers t ON t.id = g.teacher_id
		WHERE  g.id = $1
	`, groupID).Scan(&userID)
	if err == pgx.ErrNoRows {
		return uuid.Nil, ErrNotFound
	}
	if err != nil {
		return uuid.Nil, fmt.Errorf("store.GetGroupTeacherUserID: %w", err)
	}
	return userID, nil
}
