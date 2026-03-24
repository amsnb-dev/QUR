package model

import (
"time"
"github.com/google/uuid"
"encoding/json"
)

type SchoolRole struct {
ID          uuid.UUID       `json:"id"`
SchoolID    uuid.UUID       `json:"school_id"`
Name        string          `json:"name"`
Slug        string          `json:"slug"`
Permissions json.RawMessage `json:"permissions"`
IsSystem    bool            `json:"is_system"`
CreatedAt   time.Time       `json:"created_at"`
}

type SchoolStaff struct {
ID            uuid.UUID  `json:"id"`
SchoolID      uuid.UUID  `json:"school_id"`
RoleID        uuid.UUID  `json:"role_id"`
RoleName      string     `json:"role_name,omitempty"`
RoleSlug      string     `json:"role_slug,omitempty"`
FullName      string     `json:"full_name"`
Username      string     `json:"username"`
PasswordPlain *string    `json:"password_plain,omitempty"`
IsActive      bool       `json:"is_active"`
CreatedAt     time.Time  `json:"created_at"`
UpdatedAt     time.Time  `json:"updated_at"`
}

type CreateStaffRequest struct {
RoleID        string  `json:"role_id"        binding:"required"`
FullName      string  `json:"full_name"      binding:"required"`
Username      string  `json:"username"       binding:"required"`
Password      string  `json:"password"       binding:"required,min=6"`
PasswordPlain *string `json:"password_plain"`
}

type UpdateStaffRequest struct {
RoleID        *string `json:"role_id"`
FullName      *string `json:"full_name"`
Username      *string `json:"username"`
Password      *string `json:"password"`
PasswordPlain *string `json:"password_plain"`
IsActive      *bool   `json:"is_active"`
}

type SchoolSettings struct {
SchoolID   uuid.UUID `json:"school_id"`
Currency   string    `json:"currency"`
Language   string    `json:"language"`
Timezone   string    `json:"timezone"`
DateFormat string    `json:"date_format"`
LogoURL    *string   `json:"logo_url,omitempty"`
Address    *string   `json:"address,omitempty"`
Phone      *string   `json:"phone,omitempty"`
Email      *string   `json:"email,omitempty"`
UpdatedAt  time.Time `json:"updated_at"`
}

type UpdateSettingsRequest struct {
Currency   *string `json:"currency"`
Language   *string `json:"language"`
Timezone   *string `json:"timezone"`
DateFormat *string `json:"date_format"`
LogoURL    *string `json:"logo_url"`
Address    *string `json:"address"`
Phone      *string `json:"phone"`
Email      *string `json:"email"`
}
