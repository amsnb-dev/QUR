package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/quran-school/api/internal/middleware"
	"github.com/quran-school/api/internal/model"
	"github.com/quran-school/api/internal/store"
)

// StudentHandler handles /students and /student-groups endpoints.
type StudentHandler struct{}

func NewStudentHandler() *StudentHandler { return &StudentHandler{} }

// POST /students
func (h *StudentHandler) Create(c *gin.Context) {
	tx := middleware.TxFrom(c)
	user := middleware.UserFrom(c)
	schoolID, ok := resolveSchoolID(c)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "super_admin must specify a school context"})
		return
	}

	var req model.CreateStudentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	s, err := store.CreateStudent(c.Request.Context(), tx, schoolID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	_ = store.InsertAudit(c.Request.Context(), tx, store.AuditParams{
		SchoolID:  schoolPtr(schoolID),
		UserID:    &user.ID,
		Action:    "INSERT",
		TableName: "students",
		RecordID:  s.ID.String(),
		NewValues: map[string]any{"full_name": s.FullName, "status": s.Status},
	})

	c.JSON(http.StatusCreated, s)
}

// GET /students
func (h *StudentHandler) List(c *gin.Context) {
	tx := middleware.TxFrom(c)
	var groupID *uuid.UUID
	if gid := c.Query("group_id"); gid != "" {
		if parsed, err := uuid.Parse(gid); err == nil {
			groupID = &parsed
		}
	}
	f := model.ListStudentsFilter{
		IncludeArchived: c.Query("include_archived") == "1",
		Status:          c.Query("status"),
		GroupID:         groupID,
		Search:          c.Query("search"),
		Gender:          c.Query("gender"),
	}
	p := store.ParsePage(c.Query("limit"), c.Query("offset"))

	students, total, err := store.ListStudents(c.Request.Context(), tx, f, p)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if students == nil {
		students = []model.Student{}
	}
	c.JSON(http.StatusOK, store.PagedResult[model.Student]{
		Data: students,
		Meta: store.Meta{Limit: p.Limit, Offset: p.Offset, Total: total},
	})
}

// GET /students/:id
func (h *StudentHandler) Get(c *gin.Context) {
	tx := middleware.TxFrom(c)
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid student id"})
		return
	}
	s, err := store.GetStudent(c.Request.Context(), tx, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if s == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "student not found"})
		return
	}
	c.JSON(http.StatusOK, s)
}

// PUT /students/:id
func (h *StudentHandler) Update(c *gin.Context) {
	tx := middleware.TxFrom(c)
	user := middleware.UserFrom(c)
	schoolID, _ := resolveSchoolID(c)

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid student id"})
		return
	}

	var req model.UpdateStudentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	s, err := store.UpdateStudent(c.Request.Context(), tx, id, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if s == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "student not found"})
		return
	}

	nv := map[string]any{}
	if req.FullName != nil {
		nv["full_name"] = *req.FullName
	}
	if req.Status != nil {
		nv["status"] = *req.Status
	}
	if req.MemorizedParts != nil {
		nv["memorized_parts"] = *req.MemorizedParts
	}
	_ = store.InsertAudit(c.Request.Context(), tx, store.AuditParams{
		SchoolID:  schoolPtr(schoolID),
		UserID:    &user.ID,
		Action:    "UPDATE",
		TableName: "students",
		RecordID:  id.String(),
		NewValues: nv,
	})

	c.JSON(http.StatusOK, s)
}

// PATCH /students/:id/archive
func (h *StudentHandler) Archive(c *gin.Context) {
	tx := middleware.TxFrom(c)
	user := middleware.UserFrom(c)

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid student id"})
		return
	}
	schoolID, ok := resolveSchoolID(c)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "super_admin must specify a school context"})
		return
	}

	// DB function archive_student() raises EXCEPTION with Arabic message on constraint violations.
	// It also writes its own audit entry; we add an app-level entry for completeness.
	if err := store.ArchiveStudent(c.Request.Context(), tx, id, schoolID, user.ID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	_ = store.InsertAudit(c.Request.Context(), tx, store.AuditParams{
		SchoolID:  schoolPtr(schoolID),
		UserID:    &user.ID,
		Action:    "ARCHIVE",
		TableName: "students",
		RecordID:  id.String(),
		NewValues: map[string]any{"is_archived": true},
	})

	c.JSON(http.StatusOK, gin.H{"message": "student archived"})
}

// POST /students/:id/groups
func (h *StudentHandler) AddGroup(c *gin.Context) {
	tx := middleware.TxFrom(c)
	user := middleware.UserFrom(c)

	studentID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid student id"})
		return
	}
	schoolID, ok := resolveSchoolID(c)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "super_admin must specify a school context"})
		return
	}

	var req model.AddStudentGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	sg, err := store.AddStudentGroup(c.Request.Context(), tx, schoolID, studentID, &req, user.ID)
	if err != nil {
		if err == store.ErrConflict {
			c.JSON(http.StatusConflict, gin.H{"error": "student is already enrolled in this group"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	_ = store.InsertAudit(c.Request.Context(), tx, store.AuditParams{
		SchoolID:  schoolPtr(schoolID),
		UserID:    &user.ID,
		Action:    "INSERT",
		TableName: "student_groups",
		RecordID:  sg.ID.String(),
		NewValues: map[string]any{
			"student_id": studentID,
			"group_id":   req.GroupID,
			"is_primary": sg.IsPrimary,
		},
	})

	c.JSON(http.StatusCreated, sg)
}

// GET /students/:id/groups
func (h *StudentHandler) ListGroups(c *gin.Context) {
	tx := middleware.TxFrom(c)
	studentID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid student id"})
		return
	}

	schoolID, _ := resolveSchoolID(c)
	// super_admin: schoolID is uuid.Nil â€” RLS still filters via SET LOCAL app.school_id=''
	currentOnly := c.Query("current") == "1"

	memberships, err := store.ListStudentGroups(c.Request.Context(), tx, schoolID, studentID, currentOnly)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if memberships == nil {
		memberships = []model.StudentGroup{}
	}
	c.JSON(http.StatusOK, gin.H{"data": memberships, "count": len(memberships)})
}

// PATCH /student-groups/:id/close
func (h *StudentHandler) CloseGroup(c *gin.Context) {
	tx := middleware.TxFrom(c)
	user := middleware.UserFrom(c)

	membershipID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid membership id"})
		return
	}
	schoolID, ok := resolveSchoolID(c)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "super_admin must specify a school context"})
		return
	}

	var req model.CloseStudentGroupRequest
	_ = c.ShouldBindJSON(&req) // optional body

	endDate := time.Now().UTC().Truncate(24 * time.Hour)
	if req.EndDate != nil {
		endDate = *req.EndDate
	}

	sg, err := store.CloseStudentGroup(c.Request.Context(), tx, schoolID, membershipID, endDate, user.ID)
	if err != nil {
		if err == store.ErrNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "membership not found or already closed"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	_ = store.InsertAudit(c.Request.Context(), tx, store.AuditParams{
		SchoolID:  schoolPtr(schoolID),
		UserID:    &user.ID,
		Action:    "CLOSE",
		TableName: "student_groups",
		RecordID:  membershipID.String(),
		NewValues: map[string]any{"end_date": endDate.Format("2006-01-02")},
	})

	c.JSON(http.StatusOK, sg)
}


