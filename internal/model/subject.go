package model

import (
	"time"

	"github.com/google/uuid"
)

// ── Subject ────────────────────────────────────────────────

type Subject struct {
	ID          uuid.UUID  `json:"id"`
	SchoolID    uuid.UUID  `json:"school_id"`
	Name        string     `json:"name"`
	Description *string    `json:"description,omitempty"`
	Category    string     `json:"category"`             // quran|arabic|islamic|other
	Color       string     `json:"color"`
	Icon        string     `json:"icon"`
	IsActive    bool       `json:"is_active"`
	SortOrder   int        `json:"sort_order"`
	CreatedBy   *uuid.UUID `json:"created_by,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`

	// Joined
	LevelsCount int `json:"levels_count,omitempty"`
}

type CreateSubjectRequest struct {
	Name        string  `json:"name"        binding:"required"`
	Description *string `json:"description"`
	Category    string  `json:"category"    binding:"required,oneof=quran arabic islamic other"`
	Color       *string `json:"color"`
	Icon        *string `json:"icon"`
	SortOrder   *int    `json:"sort_order"`
}

type UpdateSubjectRequest struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
	Category    *string `json:"category"`
	Color       *string `json:"color"`
	Icon        *string `json:"icon"`
	IsActive    *bool   `json:"is_active"`
	SortOrder   *int    `json:"sort_order"`
}

// ── SubjectLevel ───────────────────────────────────────────

type SubjectLevel struct {
	ID          uuid.UUID `json:"id"`
	SchoolID    uuid.UUID `json:"school_id"`
	SubjectID   uuid.UUID `json:"subject_id"`
	Name        string    `json:"name"`
	Description *string   `json:"description,omitempty"`
	OrderIndex  int       `json:"order_index"`
	Criteria    *string   `json:"criteria,omitempty"`
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	// Joined
	SubjectName string `json:"subject_name,omitempty"`
}

type CreateSubjectLevelRequest struct {
	Name        string  `json:"name"        binding:"required"`
	Description *string `json:"description"`
	OrderIndex  *int    `json:"order_index"`
	Criteria    *string `json:"criteria"`
}

type UpdateSubjectLevelRequest struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
	OrderIndex  *int    `json:"order_index"`
	Criteria    *string `json:"criteria"`
	IsActive    *bool   `json:"is_active"`
}

// ── StudentSubject ─────────────────────────────────────────

type StudentSubject struct {
	ID             uuid.UUID  `json:"id"`
	SchoolID       uuid.UUID  `json:"school_id"`
	StudentID      uuid.UUID  `json:"student_id"`
	SubjectID      uuid.UUID  `json:"subject_id"`
	CurrentLevelID *uuid.UUID `json:"current_level_id,omitempty"`
	Status         string     `json:"status"` // active|paused|completed|dropped
	StartedAt      string     `json:"started_at"`
	Notes          *string    `json:"notes,omitempty"`
	AssignedBy     *uuid.UUID `json:"assigned_by,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`

	// Joined
	SubjectName      string  `json:"subject_name,omitempty"`
	SubjectIcon      string  `json:"subject_icon,omitempty"`
	SubjectColor     string  `json:"subject_color,omitempty"`
	CurrentLevelName *string `json:"current_level_name,omitempty"`
	StudentName      string  `json:"student_name,omitempty"`
}

type AssignSubjectRequest struct {
	SubjectID      uuid.UUID  `json:"subject_id"       binding:"required"`
	CurrentLevelID *uuid.UUID `json:"current_level_id"`
	StartedAt      *string    `json:"started_at"`
	Notes          *string    `json:"notes"`
}

type UpdateStudentSubjectRequest struct {
	CurrentLevelID *uuid.UUID `json:"current_level_id"`
	Status         *string    `json:"status"`
	Notes          *string    `json:"notes"`
}

// ── SubjectSession ─────────────────────────────────────────

type SubjectSession struct {
	ID              uuid.UUID  `json:"id"`
	SchoolID        uuid.UUID  `json:"school_id"`
	StudentID       uuid.UUID  `json:"student_id"`
	SubjectID       uuid.UUID  `json:"subject_id"`
	TeacherID       *uuid.UUID `json:"teacher_id,omitempty"`
	SessionDate     string     `json:"session_date"`
	Content         *string    `json:"content,omitempty"`
	PagesCount      *float64   `json:"pages_count,omitempty"`
	DurationMinutes *int       `json:"duration_minutes,omitempty"`
	Performance     *string    `json:"performance,omitempty"` // excellent|good|average|weak|absent
	Score           *float64   `json:"score,omitempty"`
	LevelID         *uuid.UUID `json:"level_id,omitempty"`
	Notes           *string    `json:"notes,omitempty"`
	RecordedBy      *uuid.UUID `json:"recorded_by,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`

	// Joined
	StudentName string  `json:"student_name,omitempty"`
	SubjectName string  `json:"subject_name,omitempty"`
	TeacherName *string `json:"teacher_name,omitempty"`
	LevelName   *string `json:"level_name,omitempty"`
}

type CreateSubjectSessionRequest struct {
	StudentID       uuid.UUID  `json:"student_id"        binding:"required"`
	SubjectID       uuid.UUID  `json:"subject_id"        binding:"required"`
	TeacherID       *uuid.UUID `json:"teacher_id"`
	SessionDate     string     `json:"session_date"`
	Content         *string    `json:"content"`
	PagesCount      *float64   `json:"pages_count"`
	DurationMinutes *int       `json:"duration_minutes"`
	Performance     *string    `json:"performance"`
	Score           *float64   `json:"score"`
	LevelID         *uuid.UUID `json:"level_id"`
	Notes           *string    `json:"notes"`
}

type ListSubjectSessionsFilter struct {
	StudentID *uuid.UUID
	SubjectID *uuid.UUID
	TeacherID *uuid.UUID
	DateFrom  *string
	DateTo    *string
}
