package model

import (
"time"
"github.com/google/uuid"
)

type Guardian struct {
ID            uuid.UUID `json:"id"`
SchoolID      uuid.UUID `json:"school_id"`
FullName      string    `json:"full_name"`
Phone         *string   `json:"phone,omitempty"`
Phone2        *string   `json:"phone2,omitempty"`
Email         *string   `json:"email,omitempty"`
Address       *string   `json:"address,omitempty"`
NationalID    *string   `json:"national_id,omitempty"`
Username      *string   `json:"username,omitempty"`
PasswordPlain *string   `json:"password_plain,omitempty"`
Notes         *string   `json:"notes,omitempty"`
Relation      *string   `json:"relation,omitempty"`
	IsActive      bool      `json:"is_active"`
CreatedAt     time.Time `json:"created_at"`
UpdatedAt     time.Time `json:"updated_at"`
}

type CreateGuardianRequest struct {
FullName      string  `json:"full_name" binding:"required,min=1,max=200"`
Phone         *string `json:"phone"`
Phone2        *string `json:"phone2"`
Email         *string `json:"email"`
Address       *string `json:"address"`
NationalID    *string `json:"national_id"`
Username      *string `json:"username"`
PasswordPlain *string `json:"password_plain"`
Notes         *string `json:"notes"`
	Relation      *string `json:"relation"`
}

type UpdateGuardianRequest struct {
FullName      *string `json:"full_name"`
Phone         *string `json:"phone"`
Phone2        *string `json:"phone2"`
Email         *string `json:"email"`
Address       *string `json:"address"`
NationalID    *string `json:"national_id"`
Username      *string `json:"username"`
PasswordPlain *string `json:"password_plain"`
Notes         *string `json:"notes"`
Relation      *string `json:"relation"`
	IsActive      *bool   `json:"is_active"`
}

type ListGuardiansFilter struct {
Search string
}
