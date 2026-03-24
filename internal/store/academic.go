package store

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/quran-school/api/internal/model"
)

// ── AcademicYear ──────────────────────────────────────────────────────────────

// CreateAcademicYear inserts a new academic year.
// If IsCurrent is true, it first clears any existing current year for the school.
func CreateAcademicYear(ctx context.Context, tx pgx.Tx, schoolID uuid.UUID, createdBy uuid.UUID, r *model.CreateAcademicYearRequest) (*model.AcademicYear, error) {
	start, err := time.Parse("2006-01-02", r.StartDate)
	if err != nil {
		return nil, fmt.Errorf("store.CreateAcademicYear parse start_date: %w", err)
	}
	end, err := time.Parse("2006-01-02", r.EndDate)
	if err != nil {
		return nil, fmt.Errorf("store.CreateAcademicYear parse end_date: %w", err)
	}
	if !end.After(start) {
		return nil, fmt.Errorf("store.CreateAcademicYear: end_date must be after start_date")
	}

	if r.IsCurrent {
		if _, err = tx.Exec(ctx, `
			UPDATE academic_years SET is_current = FALSE
			WHERE school_id = $1 AND is_current = TRUE
		`, schoolID); err != nil {
			return nil, fmt.Errorf("store.CreateAcademicYear clear current: %w", err)
		}
	}

	var ay model.AcademicYear
	err = tx.QueryRow(ctx, `
		INSERT INTO academic_years (school_id, name, start_date, end_date, is_current)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, school_id, name, start_date, end_date, is_current, created_at, updated_at
	`, schoolID, r.Name, start, end, r.IsCurrent).Scan(
		&ay.ID, &ay.SchoolID, &ay.Name, &ay.StartDate, &ay.EndDate,
		&ay.IsCurrent, &ay.CreatedAt, &ay.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("store.CreateAcademicYear insert: %w", err)
	}
	return &ay, nil
}

// ListAcademicYears returns all academic years for the current tenant.
func ListAcademicYears(ctx context.Context, tx pgx.Tx) ([]model.AcademicYear, error) {
	rows, err := tx.Query(ctx, `
		SELECT id, school_id, name, start_date, end_date, is_current, created_at, updated_at
		FROM   academic_years
		ORDER  BY start_date DESC
	`)
	if err != nil {
		return nil, fmt.Errorf("store.ListAcademicYears: %w", err)
	}
	defer rows.Close()

	var list []model.AcademicYear
	for rows.Next() {
		var ay model.AcademicYear
		if err := rows.Scan(
			&ay.ID, &ay.SchoolID, &ay.Name, &ay.StartDate, &ay.EndDate,
			&ay.IsCurrent, &ay.CreatedAt, &ay.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("store.ListAcademicYears scan: %w", err)
		}
		list = append(list, ay)
	}
	return list, rows.Err()
}

// GetAcademicYear fetches a single academic year by id.
func GetAcademicYear(ctx context.Context, tx pgx.Tx, id uuid.UUID) (*model.AcademicYear, error) {
	var ay model.AcademicYear
	err := tx.QueryRow(ctx, `
		SELECT id, school_id, name, start_date, end_date, is_current, created_at, updated_at
		FROM   academic_years
		WHERE  id = $1
	`, id).Scan(
		&ay.ID, &ay.SchoolID, &ay.Name, &ay.StartDate, &ay.EndDate,
		&ay.IsCurrent, &ay.CreatedAt, &ay.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("store.GetAcademicYear: %w", err)
	}
	return &ay, nil
}

// UpdateAcademicYear applies non-nil fields.
func UpdateAcademicYear(ctx context.Context, tx pgx.Tx, id uuid.UUID, schoolID uuid.UUID, r *model.UpdateAcademicYearRequest) (*model.AcademicYear, error) {
	cur, err := GetAcademicYear(ctx, tx, id)
	if err != nil || cur == nil {
		return cur, err
	}

	name := cur.Name
	if r.Name != nil {
		name = *r.Name
	}
	start := cur.StartDate
	if r.StartDate != nil {
		if start, err = time.Parse("2006-01-02", *r.StartDate); err != nil {
			return nil, fmt.Errorf("store.UpdateAcademicYear parse start_date: %w", err)
		}
	}
	end := cur.EndDate
	if r.EndDate != nil {
		if end, err = time.Parse("2006-01-02", *r.EndDate); err != nil {
			return nil, fmt.Errorf("store.UpdateAcademicYear parse end_date: %w", err)
		}
	}
	isCurrent := cur.IsCurrent
	if r.IsCurrent != nil {
		isCurrent = *r.IsCurrent
		if isCurrent {
			if _, err = tx.Exec(ctx, `
				UPDATE academic_years SET is_current = FALSE
				WHERE school_id = $1 AND is_current = TRUE AND id != $2
			`, schoolID, id); err != nil {
				return nil, fmt.Errorf("store.UpdateAcademicYear clear current: %w", err)
			}
		}
	}

	if _, err = tx.Exec(ctx, `
		UPDATE academic_years
		SET name = $2, start_date = $3, end_date = $4, is_current = $5
		WHERE id = $1
	`, id, name, start, end, isCurrent); err != nil {
		return nil, fmt.Errorf("store.UpdateAcademicYear: %w", err)
	}
	return GetAcademicYear(ctx, tx, id)
}

// ── StudentEnrollment ─────────────────────────────────────────────────────────

// CreateEnrollment registers a student in an academic year.
func CreateEnrollment(ctx context.Context, tx pgx.Tx, schoolID uuid.UUID, yearID uuid.UUID, enrolledBy uuid.UUID, r *model.CreateEnrollmentRequest) (*model.StudentEnrollment, error) {
	partsAtEntry := 0.0
	if r.PartsAtEntry != nil {
		partsAtEntry = *r.PartsAtEntry
	}

	var e model.StudentEnrollment
	err := tx.QueryRow(ctx, `
		INSERT INTO student_enrollments
			(school_id, student_id, academic_year_id, level_at_entry, parts_at_entry, target_parts, notes, enrolled_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, school_id, student_id, academic_year_id,
		          level_at_entry, parts_at_entry, target_parts,
		          status, notes, enrolled_by, created_at, updated_at
	`, schoolID, r.StudentID, yearID, r.LevelAtEntry, partsAtEntry, r.TargetParts, r.Notes, enrolledBy,
	).Scan(
		&e.ID, &e.SchoolID, &e.StudentID, &e.AcademicYearID,
		&e.LevelAtEntry, &e.PartsAtEntry, &e.TargetParts,
		&e.Status, &e.Notes, &e.EnrolledBy, &e.CreatedAt, &e.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("store.CreateEnrollment: %w", err)
	}
	return &e, nil
}

// ListEnrollments returns enrollments for a given academic year.
func ListEnrollments(ctx context.Context, tx pgx.Tx, yearID uuid.UUID, p Page) ([]model.StudentEnrollment, int, error) {
	var total int
	if err := tx.QueryRow(ctx, `
		SELECT COUNT(*) FROM student_enrollments WHERE academic_year_id = $1
	`, yearID).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("store.ListEnrollments count: %w", err)
	}

	rows, err := tx.Query(ctx, `
		SELECT e.id, e.school_id, e.student_id, e.academic_year_id,
		       s.full_name,
		       e.level_at_entry, e.parts_at_entry, e.target_parts,
		       e.status, e.notes, e.enrolled_by, e.created_at, e.updated_at
		FROM   student_enrollments e
		JOIN   students s ON s.id = e.student_id
		WHERE  e.academic_year_id = $1
		ORDER  BY s.full_name
		LIMIT  $2 OFFSET $3
	`, yearID, p.Limit, p.Offset)
	if err != nil {
		return nil, 0, fmt.Errorf("store.ListEnrollments: %w", err)
	}
	defer rows.Close()

	var list []model.StudentEnrollment
	for rows.Next() {
		var e model.StudentEnrollment
		if err := rows.Scan(
			&e.ID, &e.SchoolID, &e.StudentID, &e.AcademicYearID,
			&e.StudentName,
			&e.LevelAtEntry, &e.PartsAtEntry, &e.TargetParts,
			&e.Status, &e.Notes, &e.EnrolledBy, &e.CreatedAt, &e.UpdatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("store.ListEnrollments scan: %w", err)
		}
		list = append(list, e)
	}
	return list, total, rows.Err()
}

// ── Exam ──────────────────────────────────────────────────────────────────────

// CreateExam inserts a new exam.
func CreateExam(ctx context.Context, tx pgx.Tx, schoolID uuid.UUID, yearID uuid.UUID, createdBy uuid.UUID, r *model.CreateExamRequest) (*model.Exam, error) {
	examDate, err := time.Parse("2006-01-02", r.ExamDate)
	if err != nil {
		return nil, fmt.Errorf("store.CreateExam parse exam_date: %w", err)
	}

	examType := "periodic"
	if r.ExamType != "" {
		examType = r.ExamType
	}
	maxScore := 100.0
	if r.MaxScore != nil {
		maxScore = *r.MaxScore
	}
	passScore := 60.0
	if r.PassScore != nil {
		passScore = *r.PassScore
	}

	var ex model.Exam
	err = tx.QueryRow(ctx, `
		INSERT INTO exams
			(school_id, academic_year_id, group_id, name, exam_type, exam_date,
			 from_surah, to_surah, max_score, pass_score, notes, created_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		RETURNING id, school_id, academic_year_id, group_id, name, exam_type, exam_date,
		          from_surah, to_surah, max_score, pass_score, notes, created_by, created_at, updated_at
	`, schoolID, yearID, r.GroupID, r.Name, examType, examDate,
		r.FromSurah, r.ToSurah, maxScore, passScore, r.Notes, createdBy,
	).Scan(
		&ex.ID, &ex.SchoolID, &ex.AcademicYearID, &ex.GroupID, &ex.Name, &ex.ExamType, &ex.ExamDate,
		&ex.FromSurah, &ex.ToSurah, &ex.MaxScore, &ex.PassScore, &ex.Notes, &ex.CreatedBy,
		&ex.CreatedAt, &ex.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("store.CreateExam: %w", err)
	}
	return &ex, nil
}

// ListExams returns exams for a given academic year.
func ListExams(ctx context.Context, tx pgx.Tx, yearID uuid.UUID, p Page) ([]model.Exam, int, error) {
	var total int
	if err := tx.QueryRow(ctx, `
		SELECT COUNT(*) FROM exams WHERE academic_year_id = $1
	`, yearID).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("store.ListExams count: %w", err)
	}

	rows, err := tx.Query(ctx, `
		SELECT e.id, e.school_id, e.academic_year_id, e.group_id,
		       g.name AS group_name,
		       e.name, e.exam_type, e.exam_date,
		       e.from_surah, e.to_surah, e.max_score, e.pass_score,
		       e.notes, e.created_by, e.created_at, e.updated_at
		FROM   exams e
		LEFT JOIN groups g ON g.id = e.group_id
		WHERE  e.academic_year_id = $1
		ORDER  BY e.exam_date DESC
		LIMIT  $2 OFFSET $3
	`, yearID, p.Limit, p.Offset)
	if err != nil {
		return nil, 0, fmt.Errorf("store.ListExams: %w", err)
	}
	defer rows.Close()

	var list []model.Exam
	for rows.Next() {
		var ex model.Exam
		if err := rows.Scan(
			&ex.ID, &ex.SchoolID, &ex.AcademicYearID, &ex.GroupID,
			&ex.GroupName,
			&ex.Name, &ex.ExamType, &ex.ExamDate,
			&ex.FromSurah, &ex.ToSurah, &ex.MaxScore, &ex.PassScore,
			&ex.Notes, &ex.CreatedBy, &ex.CreatedAt, &ex.UpdatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("store.ListExams scan: %w", err)
		}
		list = append(list, ex)
	}
	return list, total, rows.Err()
}

// GetExam fetches a single exam by id.
func GetExam(ctx context.Context, tx pgx.Tx, id uuid.UUID) (*model.Exam, error) {
	var ex model.Exam
	err := tx.QueryRow(ctx, `
		SELECT e.id, e.school_id, e.academic_year_id, e.group_id,
		       g.name AS group_name,
		       e.name, e.exam_type, e.exam_date,
		       e.from_surah, e.to_surah, e.max_score, e.pass_score,
		       e.notes, e.created_by, e.created_at, e.updated_at
		FROM   exams e
		LEFT JOIN groups g ON g.id = e.group_id
		WHERE  e.id = $1
	`, id).Scan(
		&ex.ID, &ex.SchoolID, &ex.AcademicYearID, &ex.GroupID,
		&ex.GroupName,
		&ex.Name, &ex.ExamType, &ex.ExamDate,
		&ex.FromSurah, &ex.ToSurah, &ex.MaxScore, &ex.PassScore,
		&ex.Notes, &ex.CreatedBy, &ex.CreatedAt, &ex.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("store.GetExam: %w", err)
	}
	return &ex, nil
}

// ── ExamResult ────────────────────────────────────────────────────────────────

// BulkCreateExamResults inserts multiple results in one transaction.
func BulkCreateExamResults(ctx context.Context, tx pgx.Tx, schoolID uuid.UUID, examID uuid.UUID, recordedBy uuid.UUID, passScore float64, r *model.BulkExamResultsRequest) ([]model.ExamResult, error) {
	var list []model.ExamResult
	for _, item := range r.Items {
		var res model.ExamResult
		err := tx.QueryRow(ctx, `
			INSERT INTO exam_results
				(school_id, exam_id, student_id, score, hafz_score, tajweed_score, is_absent, notes, recorded_by)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
			ON CONFLICT (school_id, exam_id, student_id) DO UPDATE
				SET score         = EXCLUDED.score,
				    hafz_score    = EXCLUDED.hafz_score,
				    tajweed_score = EXCLUDED.tajweed_score,
				    is_absent     = EXCLUDED.is_absent,
				    notes         = EXCLUDED.notes,
				    recorded_by   = EXCLUDED.recorded_by
			RETURNING id, school_id, exam_id, student_id,
			          score, hafz_score, tajweed_score, is_absent, notes,
			          recorded_by, created_at, updated_at
		`, schoolID, examID, item.StudentID, item.Score,
			item.HafzScore, item.TajweedScore, item.IsAbsent, item.Notes, recordedBy,
		).Scan(
			&res.ID, &res.SchoolID, &res.ExamID, &res.StudentID,
			&res.Score, &res.HafzScore, &res.TajweedScore, &res.IsAbsent, &res.Notes,
			&res.RecordedBy, &res.CreatedAt, &res.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("store.BulkCreateExamResults student %s: %w", item.StudentID, err)
		}
		res.IsPassed = !res.IsAbsent && res.Score >= passScore
		list = append(list, res)
	}
	return list, nil
}

// ListExamResults returns all results for an exam with student names.
func ListExamResults(ctx context.Context, tx pgx.Tx, examID uuid.UUID) ([]model.ExamResult, error) {
	rows, err := tx.Query(ctx, `
		SELECT r.id, r.school_id, r.exam_id, r.student_id,
		       s.full_name,
		       r.score, r.hafz_score, r.tajweed_score, r.is_absent,
		       r.notes, r.recorded_by, r.created_at, r.updated_at,
		       e.pass_score
		FROM   exam_results r
		JOIN   students s ON s.id = r.student_id
		JOIN   exams    e ON e.id = r.exam_id
		WHERE  r.exam_id = $1
		ORDER  BY s.full_name
	`, examID)
	if err != nil {
		return nil, fmt.Errorf("store.ListExamResults: %w", err)
	}
	defer rows.Close()

	var list []model.ExamResult
	for rows.Next() {
		var res model.ExamResult
		var passScore float64
		if err := rows.Scan(
			&res.ID, &res.SchoolID, &res.ExamID, &res.StudentID,
			&res.StudentName,
			&res.Score, &res.HafzScore, &res.TajweedScore, &res.IsAbsent,
			&res.Notes, &res.RecordedBy, &res.CreatedAt, &res.UpdatedAt,
			&passScore,
		); err != nil {
			return nil, fmt.Errorf("store.ListExamResults scan: %w", err)
		}
		res.IsPassed = !res.IsAbsent && res.Score >= passScore
		list = append(list, res)
	}
	return list, rows.Err()
}

// ── Holiday ───────────────────────────────────────────────────────────────────

// CreateHoliday inserts a new holiday.
func CreateHoliday(ctx context.Context, tx pgx.Tx, schoolID uuid.UUID, r *model.CreateHolidayRequest) (*model.Holiday, error) {
	start, err := time.Parse("2006-01-02", r.StartDate)
	if err != nil {
		return nil, fmt.Errorf("store.CreateHoliday parse start_date: %w", err)
	}
	end, err := time.Parse("2006-01-02", r.EndDate)
	if err != nil {
		return nil, fmt.Errorf("store.CreateHoliday parse end_date: %w", err)
	}

	holidayType := "official"
	if r.HolidayType != "" {
		holidayType = r.HolidayType
	}

	var h model.Holiday
	err = tx.QueryRow(ctx, `
		INSERT INTO holidays (school_id, academic_year_id, name, start_date, end_date, holiday_type)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, school_id, academic_year_id, name, start_date, end_date, holiday_type, created_at
	`, schoolID, r.AcademicYearID, r.Name, start, end, holidayType,
	).Scan(
		&h.ID, &h.SchoolID, &h.AcademicYearID, &h.Name,
		&h.StartDate, &h.EndDate, &h.HolidayType, &h.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("store.CreateHoliday: %w", err)
	}
	return &h, nil
}

// ListHolidays returns holidays for the current tenant, optionally filtered by year.
func ListHolidays(ctx context.Context, tx pgx.Tx, yearID *uuid.UUID) ([]model.Holiday, error) {
	rows, err := tx.Query(ctx, `
		SELECT id, school_id, academic_year_id, name, start_date, end_date, holiday_type, created_at
		FROM   holidays
		WHERE  ($1::uuid IS NULL OR academic_year_id = $1)
		ORDER  BY start_date
	`, yearID)
	if err != nil {
		return nil, fmt.Errorf("store.ListHolidays: %w", err)
	}
	defer rows.Close()

	var list []model.Holiday
	for rows.Next() {
		var h model.Holiday
		if err := rows.Scan(
			&h.ID, &h.SchoolID, &h.AcademicYearID, &h.Name,
			&h.StartDate, &h.EndDate, &h.HolidayType, &h.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("store.ListHolidays scan: %w", err)
		}
		list = append(list, h)
	}
	return list, rows.Err()
}
