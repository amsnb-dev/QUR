package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/quran-school/api/internal/middleware"
	"github.com/quran-school/api/internal/model"
	"github.com/quran-school/api/internal/store"
)

type SubjectHandler struct{}

func NewSubjectHandler() *SubjectHandler { return &SubjectHandler{} }

// ── Subjects ───────────────────────────────────────────────

// POST /subjects
func (h *SubjectHandler) Create(c *gin.Context) {
	tx := middleware.TxFrom(c)
	user := middleware.UserFrom(c)
	schoolID, ok := resolveSchoolID(c)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "super_admin must specify a school context"})
		return
	}

	var req model.CreateSubjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	s, err := store.CreateSubject(c.Request.Context(), tx, schoolID, user.ID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	_ = store.InsertAudit(c.Request.Context(), tx, store.AuditParams{
		SchoolID:  schoolPtr(schoolID),
		UserID:    &user.ID,
		Action:    "INSERT",
		TableName: "subjects",
		RecordID:  s.ID.String(),
		NewValues: map[string]any{"name": s.Name, "category": s.Category},
	})

	c.JSON(http.StatusCreated, s)
}

// GET /subjects
func (h *SubjectHandler) List(c *gin.Context) {
	tx := middleware.TxFrom(c)
	includeInactive := c.Query("include_inactive") == "1"

	list, err := store.ListSubjects(c.Request.Context(), tx, includeInactive)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if list == nil {
		list = []model.Subject{}
	}
	c.JSON(http.StatusOK, list)
}

// GET /subjects/:id
func (h *SubjectHandler) Get(c *gin.Context) {
	tx := middleware.TxFrom(c)
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid subject id"})
		return
	}

	s, err := store.GetSubject(c.Request.Context(), tx, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if s == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "subject not found"})
		return
	}
	c.JSON(http.StatusOK, s)
}

// PUT /subjects/:id
func (h *SubjectHandler) Update(c *gin.Context) {
	tx := middleware.TxFrom(c)
	user := middleware.UserFrom(c)
	schoolID, _ := resolveSchoolID(c)

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid subject id"})
		return
	}

	var req model.UpdateSubjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	s, err := store.UpdateSubject(c.Request.Context(), tx, id, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if s == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "subject not found"})
		return
	}

	_ = store.InsertAudit(c.Request.Context(), tx, store.AuditParams{
		SchoolID:  schoolPtr(schoolID),
		UserID:    &user.ID,
		Action:    "UPDATE",
		TableName: "subjects",
		RecordID:  id.String(),
		NewValues: map[string]any{"name": s.Name, "is_active": s.IsActive},
	})

	c.JSON(http.StatusOK, s)
}

// ── SubjectLevels ──────────────────────────────────────────

// POST /subjects/:id/levels
func (h *SubjectHandler) CreateLevel(c *gin.Context) {
	tx := middleware.TxFrom(c)
	schoolID, ok := resolveSchoolID(c)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "super_admin must specify a school context"})
		return
	}

	subjectID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid subject id"})
		return
	}

	var req model.CreateSubjectLevelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	l, err := store.CreateSubjectLevel(c.Request.Context(), tx, schoolID, subjectID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, l)
}

// GET /subjects/:id/levels
func (h *SubjectHandler) ListLevels(c *gin.Context) {
	tx := middleware.TxFrom(c)
	subjectID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid subject id"})
		return
	}

	list, err := store.ListSubjectLevels(c.Request.Context(), tx, subjectID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if list == nil {
		list = []model.SubjectLevel{}
	}
	c.JSON(http.StatusOK, list)
}

// PUT /levels/:id
func (h *SubjectHandler) UpdateLevel(c *gin.Context) {
	tx := middleware.TxFrom(c)
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid level id"})
		return
	}

	var req model.UpdateSubjectLevelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	l, err := store.UpdateSubjectLevel(c.Request.Context(), tx, id, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if l == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "level not found"})
		return
	}
	c.JSON(http.StatusOK, l)
}

// ── StudentSubjects ────────────────────────────────────────

// POST /students/:id/subjects
func (h *SubjectHandler) AssignToStudent(c *gin.Context) {
	tx := middleware.TxFrom(c)
	user := middleware.UserFrom(c)
	schoolID, ok := resolveSchoolID(c)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "super_admin must specify a school context"})
		return
	}

	studentID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid student id"})
		return
	}

	var req model.AssignSubjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ss, err := store.AssignSubjectToStudent(c.Request.Context(), tx, schoolID, studentID, user.ID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	_ = store.InsertAudit(c.Request.Context(), tx, store.AuditParams{
		SchoolID:  schoolPtr(schoolID),
		UserID:    &user.ID,
		Action:    "INSERT",
		TableName: "student_subjects",
		RecordID:  ss.ID.String(),
		NewValues: map[string]any{"student_id": studentID, "subject_id": req.SubjectID},
	})

	c.JSON(http.StatusCreated, ss)
}

// GET /students/:id/subjects
func (h *SubjectHandler) ListStudentSubjects(c *gin.Context) {
	tx := middleware.TxFrom(c)
	studentID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid student id"})
		return
	}

	list, err := store.ListStudentSubjects(c.Request.Context(), tx, studentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if list == nil {
		list = []model.StudentSubject{}
	}
	c.JSON(http.StatusOK, list)
}

// PUT /student-subjects/:id
func (h *SubjectHandler) UpdateStudentSubject(c *gin.Context) {
	tx := middleware.TxFrom(c)
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var req model.UpdateStudentSubjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ss, err := store.UpdateStudentSubject(c.Request.Context(), tx, id, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if ss == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	c.JSON(http.StatusOK, ss)
}

// ── SubjectSessions ────────────────────────────────────────

// POST /subject-sessions
func (h *SubjectHandler) CreateSession(c *gin.Context) {
	tx := middleware.TxFrom(c)
	user := middleware.UserFrom(c)
	schoolID, ok := resolveSchoolID(c)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "super_admin must specify a school context"})
		return
	}

	var req model.CreateSubjectSessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	s, err := store.CreateSubjectSession(c.Request.Context(), tx, schoolID, user.ID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, s)
}

// GET /subject-sessions
func (h *SubjectHandler) ListSessions(c *gin.Context) {
	tx := middleware.TxFrom(c)

	f := model.ListSubjectSessionsFilter{}
	if raw := c.Query("student_id"); raw != "" {
		id, err := uuid.Parse(raw)
		if err == nil {
			f.StudentID = &id
		}
	}
	if raw := c.Query("subject_id"); raw != "" {
		id, err := uuid.Parse(raw)
		if err == nil {
			f.SubjectID = &id
		}
	}
	if raw := c.Query("teacher_id"); raw != "" {
		id, err := uuid.Parse(raw)
		if err == nil {
			f.TeacherID = &id
		}
	}
	if v := c.Query("date_from"); v != "" {
		f.DateFrom = &v
	}
	if v := c.Query("date_to"); v != "" {
		f.DateTo = &v
	}

	p := store.ParsePage(c.Query("limit"), c.Query("offset"))
	list, total, err := store.ListSubjectSessions(c.Request.Context(), tx, f, p)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if list == nil {
		list = []model.SubjectSession{}
	}
	c.JSON(http.StatusOK, store.PagedResult[model.SubjectSession]{
		Data: list,
		Meta: store.Meta{Limit: p.Limit, Offset: p.Offset, Total: total},
	})
}
