package model

import (
	"time"

	"github.com/google/uuid"
)

// ── AcademicYear ──────────────────────────────────────────────────────────────

// AcademicYear represents one school year per school.
type AcademicYear struct {
	ID        uuid.UUID `json:"id"`
	SchoolID  uuid.UUID `json:"school_id"`
	Name      string    `json:"name"`       // "2024-2025"
	StartDate time.Time `json:"start_date"`
	EndDate   time.Time `json:"end_date"`
	IsCurrent bool      `json:"is_current"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// CreateAcademicYearRequest – POST /academic-years
type CreateAcademicYearRequest struct {
	Name      string `json:"name"       binding:"required,min=1,max=20"`
	StartDate string `json:"start_date" binding:"required"` // "YYYY-MM-DD"
	EndDate   string `json:"end_date"   binding:"required"` // "YYYY-MM-DD"
	IsCurrent bool   `json:"is_current"`
}

// UpdateAcademicYearRequest – PUT /academic-years/:id
type UpdateAcademicYearRequest struct {
	Name      *string `json:"name"       binding:"omitempty,min=1,max=20"`
	StartDate *string `json:"start_date"`
	EndDate   *string `json:"end_date"`
	IsCurrent *bool   `json:"is_current"`
}

// ── StudentEnrollment ─────────────────────────────────────────────────────────

// StudentEnrollment registers a student in an academic year.
type StudentEnrollment struct {
	ID             uuid.UUID  `json:"id"`
	SchoolID       uuid.UUID  `json:"school_id"`
	StudentID      uuid.UUID  `json:"student_id"`
	AcademicYearID uuid.UUID  `json:"academic_year_id"`
	// joined fields (populated by store queries)
	StudentName    string     `json:"student_name,omitempty"`
	AcademicYear   string     `json:"academic_year,omitempty"`
	LevelAtEntry   *string    `json:"level_at_entry,omitempty"`
	PartsAtEntry   float64    `json:"parts_at_entry"`
	TargetParts    *float64   `json:"target_parts,omitempty"`
	Status         string     `json:"status"`
	Notes          *string    `json:"notes,omitempty"`
	EnrolledBy     uuid.UUID  `json:"enrolled_by"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

// CreateEnrollmentRequest – POST /academic-years/:id/enrollments
type CreateEnrollmentRequest struct {
	StudentID    uuid.UUID `json:"student_id"   binding:"required"`
	LevelAtEntry *string   `json:"level_at_entry"`
	PartsAtEntry *float64  `json:"parts_at_entry" binding:"omitempty,min=0,max=30"`
	TargetParts  *float64  `json:"target_parts"   binding:"omitempty,min=0,max=30"`
	Notes        *string   `json:"notes"`
}

// UpdateEnrollmentRequest – PATCH /enrollments/:id
type UpdateEnrollmentRequest struct {
	TargetParts *float64 `json:"target_parts" binding:"omitempty,min=0,max=30"`
	Status      *string  `json:"status"       binding:"omitempty,oneof=active completed withdrawn"`
	Notes       *string  `json:"notes"`
}

// ── Exam ──────────────────────────────────────────────────────────────────────

// Exam is a periodic or final exam for a school/group.
type Exam struct {
	ID             uuid.UUID  `json:"id"`
	SchoolID       uuid.UUID  `json:"school_id"`
	AcademicYearID uuid.UUID  `json:"academic_year_id"`
	GroupID        *uuid.UUID `json:"group_id,omitempty"`
	// joined fields
	GroupName      *string    `json:"group_name,omitempty"`
	Name           string     `json:"name"`
	ExamType       string     `json:"exam_type"`
	ExamDate       time.Time  `json:"exam_date"`
	FromSurah      *int16     `json:"from_surah,omitempty"`
	ToSurah        *int16     `json:"to_surah,omitempty"`
	MaxScore       float64    `json:"max_score"`
	PassScore      float64    `json:"pass_score"`
	Notes          *string    `json:"notes,omitempty"`
	CreatedBy      uuid.UUID  `json:"created_by"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

// CreateExamRequest – POST /academic-years/:id/exams
type CreateExamRequest struct {
	GroupID   *uuid.UUID `json:"group_id"`
	Name      string     `json:"name"       binding:"required,min=1,max=200"`
	ExamType  string     `json:"exam_type"  binding:"omitempty,oneof=periodic final placement competition"`
	ExamDate  string     `json:"exam_date"  binding:"required"` // "YYYY-MM-DD"
	FromSurah *int16     `json:"from_surah" binding:"omitempty,min=1,max=114"`
	ToSurah   *int16     `json:"to_surah"   binding:"omitempty,min=1,max=114"`
	MaxScore  *float64   `json:"max_score"  binding:"omitempty,min=1"`
	PassScore *float64   `json:"pass_score" binding:"omitempty,min=0"`
	Notes     *string    `json:"notes"`
}

// UpdateExamRequest – PUT /exams/:id
type UpdateExamRequest struct {
	Name      *string  `json:"name"       binding:"omitempty,min=1,max=200"`
	ExamType  *string  `json:"exam_type"  binding:"omitempty,oneof=periodic final placement competition"`
	ExamDate  *string  `json:"exam_date"`
	FromSurah *int16   `json:"from_surah" binding:"omitempty,min=1,max=114"`
	ToSurah   *int16   `json:"to_surah"   binding:"omitempty,min=1,max=114"`
	MaxScore  *float64 `json:"max_score"  binding:"omitempty,min=1"`
	PassScore *float64 `json:"pass_score" binding:"omitempty,min=0"`
	Notes     *string  `json:"notes"`
}

// ── ExamResult ────────────────────────────────────────────────────────────────

// ExamResult is one student's result in an exam.
type ExamResult struct {
	ID           uuid.UUID `json:"id"`
	SchoolID     uuid.UUID `json:"school_id"`
	ExamID       uuid.UUID `json:"exam_id"`
	StudentID    uuid.UUID `json:"student_id"`
	// joined fields
	StudentName  string    `json:"student_name,omitempty"`
	Score        float64   `json:"score"`
	HafzScore    *float64  `json:"hafz_score,omitempty"`
	TajweedScore *float64  `json:"tajweed_score,omitempty"`
	IsPassed     bool      `json:"is_passed"` // computed: score >= exam.pass_score
	IsAbsent     bool      `json:"is_absent"`
	Notes        *string   `json:"notes,omitempty"`
	RecordedBy   uuid.UUID `json:"recorded_by"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// ExamResultItem is one student entry inside a bulk result request.
type ExamResultItem struct {
	StudentID    uuid.UUID `json:"student_id"    binding:"required"`
	Score        float64   `json:"score"         binding:"required,min=0"`
	HafzScore    *float64  `json:"hafz_score"    binding:"omitempty,min=0"`
	TajweedScore *float64  `json:"tajweed_score" binding:"omitempty,min=0"`
	IsAbsent     bool      `json:"is_absent"`
	Notes        *string   `json:"notes"`
}

// BulkExamResultsRequest – POST /exams/:id/results
type BulkExamResultsRequest struct {
	Items []ExamResultItem `json:"items" binding:"required,min=1"`
}

// UpdateExamResultRequest – PATCH /exam-results/:id
type UpdateExamResultRequest struct {
	Score        *float64 `json:"score"         binding:"omitempty,min=0"`
	HafzScore    *float64 `json:"hafz_score"    binding:"omitempty,min=0"`
	TajweedScore *float64 `json:"tajweed_score" binding:"omitempty,min=0"`
	IsAbsent     *bool    `json:"is_absent"`
	Notes        *string  `json:"notes"`
}

// ── Holiday ───────────────────────────────────────────────────────────────────

// Holiday is a school holiday or official break.
type Holiday struct {
	ID             uuid.UUID  `json:"id"`
	SchoolID       uuid.UUID  `json:"school_id"`
	AcademicYearID *uuid.UUID `json:"academic_year_id,omitempty"`
	Name           string     `json:"name"`
	StartDate      time.Time  `json:"start_date"`
	EndDate        time.Time  `json:"end_date"`
	HolidayType    string     `json:"holiday_type"`
	CreatedAt      time.Time  `json:"created_at"`
}

// CreateHolidayRequest – POST /holidays
type CreateHolidayRequest struct {
	AcademicYearID *uuid.UUID `json:"academic_year_id"`
	Name           string     `json:"name"         binding:"required,min=1,max=200"`
	StartDate      string     `json:"start_date"   binding:"required"` // "YYYY-MM-DD"
	EndDate        string     `json:"end_date"     binding:"required"` // "YYYY-MM-DD"
	HolidayType    string     `json:"holiday_type" binding:"omitempty,oneof=official school emergency"`
}
