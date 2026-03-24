package model

import "github.com/google/uuid"

// Role names — must match the roles table seed data
const (
	RoleSuperAdmin  = "super_admin"
	RoleSchoolAdmin = "school_admin"
	RoleSupervisor  = "supervisor"
	RoleTeacher     = "teacher"
	RoleAccountant  = "accountant"
)

// User is the authenticated principal, built from JWT claims or DB.
type User struct {
	ID       uuid.UUID  `json:"id"`
	SchoolID *uuid.UUID `json:"school_id"` // nil for super_admin
	Role     string     `json:"role"`
	FullName string     `json:"full_name"`
	Email    string     `json:"email"`
}

// IsSuperAdmin returns true when the user has platform-wide access.
func (u *User) IsSuperAdmin() bool { return u.Role == RoleSuperAdmin }

// AccessClaims are the custom fields embedded inside the JWT access token.
type AccessClaims struct {
	UserID   string  `json:"uid"`
	SchoolID *string `json:"sid"` // nil → super_admin
	Role     string  `json:"role"`
}

// LoginRequest is the JSON body for POST /auth/login.
type LoginRequest struct {
	Email    string `json:"email"    binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

// RefreshRequest is the JSON body for POST /auth/refresh.
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// TokenPair is returned on successful login or refresh.
type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"` // seconds until access token expires
}
