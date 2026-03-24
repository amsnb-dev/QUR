package model

import (
	"time"

	"github.com/google/uuid"
)

// â”€â”€ Group â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€

type Group struct {
	ID         uuid.UUID  `json:"id"`
	SchoolID   uuid.UUID  `json:"school_id"`
	TeacherID  *uuid.UUID `json:"teacher_id,omitempty"`
	Name       string     `json:"name"`
	Level      *string    `json:"level,omitempty"`
	Stage      *string    `json:"stage,omitempty"` // Ø§Ù„Ù…Ø±Ø­Ù„Ø©: primary/middle/secondary/adult/custom
	Room       *string    `json:"room,omitempty"`
	Days       []int16    `json:"days"`
	StartTime  *string    `json:"start_time,omitempty"`
	EndTime    *string    `json:"end_time,omitempty"`
	IsActive   bool       `json:"is_active"`
	IsArchived bool       `json:"is_archived"`
	ArchivedAt *time.Time `json:"archived_at,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`

	// Joined fields
	TeacherName   *string `json:"teacher_name,omitempty"`
	StudentsCount int     `json:"students_count,omitempty"`
}

// CreateGroupRequest is the JSON body for POST /groups.
type CreateGroupRequest struct {
	TeacherID *uuid.UUID `json:"teacher_id"` // Ø§Ø®ØªÙŠØ§Ø±ÙŠ
	Name      string     `json:"name"        binding:"required,min=1,max=120"`
	Level     *string    `json:"level"`
	Stage     *string    `json:"stage"`
	Room      *string    `json:"room"`
	Days      []int16    `json:"days"`
	StartTime *string    `json:"start_time"`
	EndTime   *string    `json:"end_time"`
}

// UpdateGroupRequest is the JSON body for PUT /groups/:id.
type UpdateGroupRequest struct {
	TeacherID *uuid.UUID `json:"teacher_id"`
	Name      *string    `json:"name"      binding:"omitempty,min=1,max=120"`
	Level     *string    `json:"level"`
	Stage     *string    `json:"stage"`
	Room      *string    `json:"room"`
	Days      []int16    `json:"days"`
	StartTime *string    `json:"start_time"`
	EndTime   *string    `json:"end_time"`
	IsActive  *bool      `json:"is_active"`
}

// ListGroupsFilter holds query-string filters for GET /groups.
type ListGroupsFilter struct {
	IncludeArchived bool
}

// â”€â”€ Student â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€

type Student struct {
	ID             uuid.UUID  `json:"id"`
	SchoolID       uuid.UUID  `json:"school_id"`
	FullName       string     `json:"full_name"`
	DateOfBirth    *time.Time `json:"date_of_birth,omitempty"`
	GuardianName   *string    `json:"guardian_name,omitempty"`
	EnrollmentDate time.Time  `json:"enrollment_date"`
	MemorizedParts float64    `json:"memorized_parts"`
	LevelOnEntry   *string    `json:"level_on_entry,omitempty"`
	MonthlyFee     *float64   `json:"monthly_fee,omitempty"`
	FeeExemption   string     `json:"fee_exemption"`
	Status         string     `json:"status"`
	Notes          *string    `json:"notes,omitempty"`
	Gender         *string    `json:"gender,omitempty"`
	Username       *string    `json:"username,omitempty"`
	PasswordPlain  *string    `json:"password_plain,omitempty"`
	GuardianID     *string    `json:"guardian_id,omitempty"`
	IsArchived     bool       `json:"is_archived"`
	ArchivedAt     *time.Time `json:"archived_at,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

type CreateStudentRequest struct {
	FullName       string     `json:"full_name"     binding:"required,min=1,max=200"`
	DateOfBirth    *time.Time `json:"date_of_birth"`
	GuardianName   *string    `json:"guardian_name"`
	EnrollmentDate *time.Time `json:"enrollment_date"`
	MemorizedParts *float64   `json:"memorized_parts"`
	LevelOnEntry   *string    `json:"level_on_entry"`
	MonthlyFee     *float64   `json:"monthly_fee"`
	FeeExemption   *string    `json:"fee_exemption"`
	Notes          *string    `json:"notes"`
	Gender         *string    `json:"gender"`
	Username       *string    `json:"username"`
	PasswordPlain  *string    `json:"password_plain"`
	GuardianID     *string    `json:"guardian_id"`
}

type UpdateStudentRequest struct {
	FullName       *string    `json:"full_name"     binding:"omitempty,min=1,max=200"`
	DateOfBirth    *time.Time `json:"date_of_birth"`
	GuardianName   *string    `json:"guardian_name"`
	MemorizedParts *float64   `json:"memorized_parts"`
	LevelOnEntry   *string    `json:"level_on_entry"`
	MonthlyFee     *float64   `json:"monthly_fee"`
	FeeExemption   *string    `json:"fee_exemption"`
	Status         *string    `json:"status"`
	Notes          *string    `json:"notes"`
	Gender         *string    `json:"gender"`
	Username       *string    `json:"username"`
	PasswordPlain  *string    `json:"password_plain"`
	GuardianID     *string    `json:"guardian_id"`
}

type ListStudentsFilter struct {
IncludeArchived bool
Status          string
GroupID         *uuid.UUID
Search          string
Gender          string
}

// â”€â”€ StudentGroup â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€

type StudentGroup struct {
	ID        uuid.UUID  `json:"id"`
	SchoolID  uuid.UUID  `json:"school_id"`
	StudentID uuid.UUID  `json:"student_id"`
	GroupID   uuid.UUID  `json:"group_id"`
	GroupName *string    `json:"group_name,omitempty"`
	StartDate time.Time  `json:"start_date"`
	EndDate   *time.Time `json:"end_date,omitempty"`
	IsPrimary bool       `json:"is_primary"`
	IsActive  bool       `json:"is_active"`
	CreatedAt time.Time  `json:"created_at"`
}

type AddStudentGroupRequest struct {
	GroupID   uuid.UUID  `json:"group_id"   binding:"required"`
	StartDate *time.Time `json:"start_date"`
	IsPrimary *bool      `json:"is_primary"`
}

type CloseStudentGroupRequest struct {
	EndDate *time.Time `json:"end_date"`
}

