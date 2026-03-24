package model

import (
	"time"

	"github.com/google/uuid"
)

// ── Attendance ────────────────────────────────────────────────

// AttendanceRecord is one row in the attendance table.
type AttendanceRecord struct {
	ID         uuid.UUID  `json:"id"`
	SchoolID   uuid.UUID  `json:"school_id"`
	StudentID  uuid.UUID  `json:"student_id"`
	GroupID    *uuid.UUID `json:"group_id"`
	RecordedBy uuid.UUID  `json:"recorded_by"`
	Date       time.Time  `json:"date"`
	Status     string     `json:"status"`
	Note       *string    `json:"note,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
}

// AttendanceItem is one student entry inside a batch attendance request.
type AttendanceItem struct {
	StudentID uuid.UUID `json:"student_id" binding:"required"`
	Status    string    `json:"status"     binding:"required,oneof=present absent late excused"`
	Note      *string   `json:"note"`
}

// BulkAttendanceRequest is the body for POST /groups/:id/attendance.
type BulkAttendanceRequest struct {
	Date  string           `json:"date"  binding:"required"` // "YYYY-MM-DD"
	Items []AttendanceItem `json:"items" binding:"required,min=1"`
}

// UpdateAttendanceRequest is the body for PATCH /attendance/:id.
type UpdateAttendanceRequest struct {
	Status *string `json:"status" binding:"omitempty,oneof=present absent late excused"`
	Note   *string `json:"note"`
}

// ── Memorization ──────────────────────────────────────────────

// MemorizationRecord is one row in memorization_logs.
type MemorizationRecord struct {
	ID          uuid.UUID  `json:"id"`
	SchoolID    uuid.UUID  `json:"school_id"`
	StudentID   uuid.UUID  `json:"student_id"`
	GroupID     *uuid.UUID `json:"group_id,omitempty"`
	RecordedBy  uuid.UUID  `json:"recorded_by"`
	Date        time.Time  `json:"date"`
	SurahNumber int16      `json:"surah_number"`
	FromVerse   int16      `json:"from_verse"`
	ToVerse     int16      `json:"to_verse"`
	EntryType   string     `json:"entry_type"`
	Grade       int16      `json:"grade"`
	Notes       *string    `json:"notes,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
}

// CreateMemorizationRequest is the body for POST /students/:id/memorization.
type CreateMemorizationRequest struct {
	Date        time.Time  `json:"date"         binding:"required"`
	SurahNumber int16      `json:"surah_number" binding:"required,min=1,max=114"`
	FromVerse   int16      `json:"from_verse"   binding:"required,min=1"`
	ToVerse     int16      `json:"to_verse"     binding:"required,min=1"`
	EntryType   string     `json:"entry_type"   binding:"omitempty,oneof=new review"`
	Grade       int16      `json:"grade"        binding:"required,min=1,max=5"`
	Notes       *string    `json:"notes"`
	GroupID     *uuid.UUID `json:"group_id"` // optional, validated server-side
}

// UpdateMemorizationRequest is the body for PATCH /memorization/:id.
type UpdateMemorizationRequest struct {
	Grade     *int16  `json:"grade"      binding:"omitempty,min=1,max=5"`
	Notes     *string `json:"notes"`
	FromVerse *int16  `json:"from_verse" binding:"omitempty,min=1"`
	ToVerse   *int16  `json:"to_verse"   binding:"omitempty,min=1"`
}
