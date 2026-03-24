package model

import (
	"time"

	"github.com/google/uuid"
)

// UserRecord represents a system user row from the DB.
type UserRecord struct {
	ID        uuid.UUID  `json:"id"`
	SchoolID  *uuid.UUID `json:"school_id,omitempty"`
	RoleID    int16      `json:"role_id"`
	RoleName  string     `json:"role"`
	FullName  string     `json:"full_name"`
	Email     string     `json:"email"`
	IsActive  bool       `json:"is_active"`
	CreatedAt time.Time  `json:"created_at"`
}

type CreateUserRequest struct {
	FullName string  `json:"full_name" binding:"required"`
	Email    string  `json:"email"     binding:"required,email"`
	Password string  `json:"password"  binding:"required,min=8"`
	Role     string  `json:"role"      binding:"required"`
	SchoolID *string `json:"school_id,omitempty"`
}

type UpdateUserRequest struct {
	FullName *string `json:"full_name"`
	Email    *string `json:"email"`
	Password *string `json:"password"`
	IsActive *bool   `json:"is_active"`
}

type ListUsersFilter struct {
	IncludeArchived bool
	RoleName        string
}
