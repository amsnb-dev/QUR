package model

import (
	"time"

	"github.com/google/uuid"
)

type School struct {
	ID                uuid.UUID `json:"id"`
	Name              string    `json:"name"`
	City              *string   `json:"city,omitempty"`
	Country           string    `json:"country"`
	Plan              string    `json:"plan"`
	IsActive          bool      `json:"is_active"`
	DefaultMonthlyFee float64   `json:"default_monthly_fee"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
	StudentCount      int       `json:"student_count"`
	TeacherCount      int       `json:"teacher_count"`
	GroupCount        int       `json:"group_count"`
}

type CreateSchoolRequest struct {
	Name              string   `json:"name"               binding:"required,min=1,max=200"`
	City              *string  `json:"city"`
	Country           *string  `json:"country"`
	Plan              *string  `json:"plan"`
	DefaultMonthlyFee *float64 `json:"default_monthly_fee"`
}

type UpdateSchoolRequest struct {
	Name              *string  `json:"name"               binding:"omitempty,min=1,max=200"`
	City              *string  `json:"city"`
	Country           *string  `json:"country"`
	Plan              *string  `json:"plan"`
	IsActive          *bool    `json:"is_active"`
	DefaultMonthlyFee *float64 `json:"default_monthly_fee"`
}
