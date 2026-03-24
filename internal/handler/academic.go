package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/quran-school/api/internal/middleware"
	"github.com/quran-school/api/internal/model"
	"github.com/quran-school/api/internal/store"
)

// AcademicHandler handles /academic-years and related endpoints.
type AcademicHandler struct{}

func NewAcademicHandler() *AcademicHandler { return &AcademicHandler{} }

// ── AcademicYear ──────────────────────────────────────────────────────────────

// POST /academic-years
func (h *AcademicHandler) CreateYear(c *gin.Context) {
	tx := middleware.TxFrom(c)
	user := middleware.UserFrom(c)
	schoolID, ok := resolveSchoolID(c)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "super_admin must specify a school context"})
		return
	}

	var req model.CreateAcademicYearRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ay, err := store.CreateAcademicYear(c.Request.Context(), tx, schoolID, user.ID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	_ = store.InsertAudit(c.Request.Context(), tx, store.AuditParams{
		SchoolID:  schoolPtr(schoolID),
		UserID:    &user.ID,
		Action:    "INSERT",
		TableName: "academic_years",
		RecordID:  ay.ID.String(),
		NewValues: map[string]any{"name": ay.Name, "is_current": ay.IsCurrent},
	})

	c.JSON(http.StatusCreated, ay)
}

// GET /academic-years
func (h *AcademicHandler) ListYears(c *gin.Context) {
	tx := middleware.TxFrom(c)

	list, err := store.ListAcademicYears(c.Request.Context(), tx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if list == nil {
		list = []model.AcademicYear{}
	}
	c.JSON(http.StatusOK, list)
}

// GET /academic-years/:id
func (h *AcademicHandler) GetYear(c *gin.Context) {
	tx := middleware.TxFrom(c)
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid academic year id"})
		return
	}

	ay, err := store.GetAcademicYear(c.Request.Context(), tx, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if ay == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "academic year not found"})
		return
	}
	c.JSON(http.StatusOK, ay)
}

// PUT /academic-years/:id
func (h *AcademicHandler) UpdateYear(c *gin.Context) {
	tx := middleware.TxFrom(c)
	user := middleware.UserFrom(c)
	schoolID, _ := resolveSchoolID(c)

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid academic year id"})
		return
	}

	var req model.UpdateAcademicYearRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ay, err := store.UpdateAcademicYear(c.Request.Context(), tx, id, schoolID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if ay == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "academic year not found"})
		return
	}

	_ = store.InsertAudit(c.Request.Context(), tx, store.AuditParams{
		SchoolID:  schoolPtr(schoolID),
		UserID:    &user.ID,
		Action:    "UPDATE",
		TableName: "academic_years",
		RecordID:  id.String(),
		NewValues: map[string]any{"name": ay.Name, "is_current": ay.IsCurrent},
	})

	c.JSON(http.StatusOK, ay)
}

// ── StudentEnrollment ─────────────────────────────────────────────────────────

// POST /academic-years/:id/enrollments
func (h *AcademicHandler) CreateEnrollment(c *gin.Context) {
	tx := middleware.TxFrom(c)
	user := middleware.UserFrom(c)
	schoolID, ok := resolveSchoolID(c)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "super_admin must specify a school context"})
		return
	}

	yearID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid academic year id"})
		return
	}

	var req model.CreateEnrollmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	e, err := store.CreateEnrollment(c.Request.Context(), tx, schoolID, yearID, user.ID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	_ = store.InsertAudit(c.Request.Context(), tx, store.AuditParams{
		SchoolID:  schoolPtr(schoolID),
		UserID:    &user.ID,
		Action:    "INSERT",
		TableName: "student_enrollments",
		RecordID:  e.ID.String(),
		NewValues: map[string]any{"student_id": req.StudentID, "academic_year_id": yearID},
	})

	c.JSON(http.StatusCreated, e)
}

// GET /academic-years/:id/enrollments
func (h *AcademicHandler) ListEnrollments(c *gin.Context) {
	tx := middleware.TxFrom(c)
	yearID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid academic year id"})
		return
	}

	p := store.ParsePage(c.Query("limit"), c.Query("offset"))
	list, total, err := store.ListEnrollments(c.Request.Context(), tx, yearID, p)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if list == nil {
		list = []model.StudentEnrollment{}
	}
	c.JSON(http.StatusOK, store.PagedResult[model.StudentEnrollment]{
		Data: list,
		Meta: store.Meta{Limit: p.Limit, Offset: p.Offset, Total: total},
	})
}

// ── Exam ──────────────────────────────────────────────────────────────────────

// POST /academic-years/:id/exams
func (h *AcademicHandler) CreateExam(c *gin.Context) {
	tx := middleware.TxFrom(c)
	user := middleware.UserFrom(c)
	schoolID, ok := resolveSchoolID(c)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "super_admin must specify a school context"})
		return
	}

	yearID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid academic year id"})
		return
	}

	var req model.CreateExamRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ex, err := store.CreateExam(c.Request.Context(), tx, schoolID, yearID, user.ID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	_ = store.InsertAudit(c.Request.Context(), tx, store.AuditParams{
		SchoolID:  schoolPtr(schoolID),
		UserID:    &user.ID,
		Action:    "INSERT",
		TableName: "exams",
		RecordID:  ex.ID.String(),
		NewValues: map[string]any{"name": ex.Name, "exam_type": ex.ExamType, "exam_date": ex.ExamDate},
	})

	c.JSON(http.StatusCreated, ex)
}

// GET /academic-years/:id/exams
func (h *AcademicHandler) ListExams(c *gin.Context) {
	tx := middleware.TxFrom(c)
	yearID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid academic year id"})
		return
	}

	p := store.ParsePage(c.Query("limit"), c.Query("offset"))
	list, total, err := store.ListExams(c.Request.Context(), tx, yearID, p)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if list == nil {
		list = []model.Exam{}
	}
	c.JSON(http.StatusOK, store.PagedResult[model.Exam]{
		Data: list,
		Meta: store.Meta{Limit: p.Limit, Offset: p.Offset, Total: total},
	})
}

// GET /exams/:id
func (h *AcademicHandler) GetExam(c *gin.Context) {
	tx := middleware.TxFrom(c)
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid exam id"})
		return
	}

	ex, err := store.GetExam(c.Request.Context(), tx, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if ex == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "exam not found"})
		return
	}
	c.JSON(http.StatusOK, ex)
}

// ── ExamResult ────────────────────────────────────────────────────────────────

// POST /exams/:id/results
func (h *AcademicHandler) BulkCreateResults(c *gin.Context) {
	tx := middleware.TxFrom(c)
	user := middleware.UserFrom(c)
	schoolID, ok := resolveSchoolID(c)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "super_admin must specify a school context"})
		return
	}

	examID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid exam id"})
		return
	}

	// fetch exam to get pass_score
	ex, err := store.GetExam(c.Request.Context(), tx, examID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if ex == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "exam not found"})
		return
	}

	var req model.BulkExamResultsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	results, err := store.BulkCreateExamResults(c.Request.Context(), tx, schoolID, examID, user.ID, ex.PassScore, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	_ = store.InsertAudit(c.Request.Context(), tx, store.AuditParams{
		SchoolID:  schoolPtr(schoolID),
		UserID:    &user.ID,
		Action:    "INSERT",
		TableName: "exam_results",
		RecordID:  examID.String(),
		NewValues: map[string]any{"count": len(results)},
	})

	c.JSON(http.StatusCreated, results)
}

// GET /exams/:id/results
func (h *AcademicHandler) ListResults(c *gin.Context) {
	tx := middleware.TxFrom(c)
	examID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid exam id"})
		return
	}

	list, err := store.ListExamResults(c.Request.Context(), tx, examID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if list == nil {
		list = []model.ExamResult{}
	}
	c.JSON(http.StatusOK, list)
}

// ── Holiday ───────────────────────────────────────────────────────────────────

// POST /holidays
func (h *AcademicHandler) CreateHoliday(c *gin.Context) {
	tx := middleware.TxFrom(c)
	user := middleware.UserFrom(c)
	schoolID, ok := resolveSchoolID(c)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "super_admin must specify a school context"})
		return
	}

	var req model.CreateHolidayRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	hol, err := store.CreateHoliday(c.Request.Context(), tx, schoolID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	_ = store.InsertAudit(c.Request.Context(), tx, store.AuditParams{
		SchoolID:  schoolPtr(schoolID),
		UserID:    &user.ID,
		Action:    "INSERT",
		TableName: "holidays",
		RecordID:  hol.ID.String(),
		NewValues: map[string]any{"name": hol.Name, "start_date": hol.StartDate},
	})

	c.JSON(http.StatusCreated, hol)
}

// GET /holidays
func (h *AcademicHandler) ListHolidays(c *gin.Context) {
	tx := middleware.TxFrom(c)

	var yearID *uuid.UUID
	if raw := c.Query("academic_year_id"); raw != "" {
		id, err := uuid.Parse(raw)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid academic_year_id"})
			return
		}
		yearID = &id
	}

	list, err := store.ListHolidays(c.Request.Context(), tx, yearID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if list == nil {
		list = []model.Holiday{}
	}
	c.JSON(http.StatusOK, list)
}
