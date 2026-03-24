package model

import (
	"time"

	"github.com/google/uuid"
)

// Teacher represents a teacher record linked to a user account.
type Teacher struct {
	ID             uuid.UUID  `json:"id"`
	SchoolID       uuid.UUID  `json:"school_id"`
	UserID         uuid.UUID  `json:"user_id"`
	FullName       string     `json:"full_name"`
	Email          string     `json:"email"`
	Specialization *string    `json:"specialization,omitempty"`
	Qualification  *string    `json:"qualification,omitempty"`
	HireDate       *time.Time `json:"hire_date,omitempty"`
	BaseSalary     float64    `json:"base_salary"`
	IsActive       bool       `json:"is_active"`
	IsArchived     bool       `json:"is_archived"`
	ArchivedAt     *time.Time `json:"archived_at,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

// CreateTeacherRequest — POST /teachers
// ينشئ user + teacher في نفس الـ transaction
type CreateTeacherRequest struct {
	FullName       string     `json:"full_name"       binding:"required,min=1,max=200"`
	Email          string     `json:"email"           binding:"required,email"`
	Password       string     `json:"password"        binding:"required,min=8"`
	Specialization *string    `json:"specialization"`
	Qualification  *string    `json:"qualification"`
	HireDate       *time.Time `json:"hire_date"`
	BaseSalary     *float64   `json:"base_salary"`
}

// UpdateTeacherRequest — PUT /teachers/:id
type UpdateTeacherRequest struct {
	FullName       *string    `json:"full_name"       binding:"omitempty,min=1,max=200"`
	Specialization *string    `json:"specialization"`
	Qualification  *string    `json:"qualification"`
	HireDate       *time.Time `json:"hire_date"`
	BaseSalary     *float64   `json:"base_salary"`
	IsActive       *bool      `json:"is_active"`
}

// ListTeachersFilter — GET /teachers
type ListTeachersFilter struct {
	IncludeArchived bool
}
