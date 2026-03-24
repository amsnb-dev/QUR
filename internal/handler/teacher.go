package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/quran-school/api/internal/middleware"
	"github.com/quran-school/api/internal/model"
	"github.com/quran-school/api/internal/store"
)

// TeacherHandler handles /teachers endpoints.
type TeacherHandler struct{}

func NewTeacherHandler() *TeacherHandler { return &TeacherHandler{} }

// POST /teachers
func (h *TeacherHandler) Create(c *gin.Context) {
	tx := middleware.TxFrom(c)
	user := middleware.UserFrom(c)
	schoolID, ok := resolveSchoolID(c)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "super_admin must specify a school context"})
		return
	}

	var req model.CreateTeacherRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	roleID, err := store.GetTeacherRoleID(c.Request.Context(), tx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not resolve teacher role"})
		return
	}

	t, err := store.CreateTeacher(c.Request.Context(), tx, schoolID, roleID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	_ = store.InsertAudit(c.Request.Context(), tx, store.AuditParams{
		SchoolID:  schoolPtr(schoolID),
		UserID:    &user.ID,
		Action:    "INSERT",
		TableName: "teachers",
		RecordID:  t.ID.String(),
		NewValues: map[string]any{"full_name": t.FullName, "email": t.Email},
	})

	c.JSON(http.StatusCreated, t)
}

// GET /teachers
func (h *TeacherHandler) List(c *gin.Context) {
	tx := middleware.TxFrom(c)
	f := model.ListTeachersFilter{IncludeArchived: c.Query("include_archived") == "1"}
	p := store.ParsePage(c.Query("limit"), c.Query("offset"))

	teachers, total, err := store.ListTeachers(c.Request.Context(), tx, f, p)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if teachers == nil {
		teachers = []model.Teacher{}
	}
	c.JSON(http.StatusOK, store.PagedResult[model.Teacher]{
		Data: teachers,
		Meta: store.Meta{Limit: p.Limit, Offset: p.Offset, Total: total},
	})
}

// GET /teachers/:id
func (h *TeacherHandler) Get(c *gin.Context) {
	tx := middleware.TxFrom(c)
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid teacher id"})
		return
	}
	t, err := store.GetTeacher(c.Request.Context(), tx, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if t == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "teacher not found"})
		return
	}
	c.JSON(http.StatusOK, t)
}

// PUT /teachers/:id
func (h *TeacherHandler) Update(c *gin.Context) {
	tx := middleware.TxFrom(c)
	user := middleware.UserFrom(c)
	schoolID, _ := resolveSchoolID(c)

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid teacher id"})
		return
	}

	var req model.UpdateTeacherRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	t, err := store.UpdateTeacher(c.Request.Context(), tx, id, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if t == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "teacher not found"})
		return
	}

	_ = store.InsertAudit(c.Request.Context(), tx, store.AuditParams{
		SchoolID:  schoolPtr(schoolID),
		UserID:    &user.ID,
		Action:    "UPDATE",
		TableName: "teachers",
		RecordID:  id.String(),
		NewValues: map[string]any{"full_name": t.FullName, "is_active": t.IsActive},
	})

	c.JSON(http.StatusOK, t)
}

// PATCH /teachers/:id/archive
func (h *TeacherHandler) Archive(c *gin.Context) {
	tx := middleware.TxFrom(c)
	user := middleware.UserFrom(c)
	schoolID, _ := resolveSchoolID(c)

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid teacher id"})
		return
	}

	if err := store.ArchiveTeacher(c.Request.Context(), tx, id); err != nil {
		if err == store.ErrNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "teacher not found or already archived"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	_ = store.InsertAudit(c.Request.Context(), tx, store.AuditParams{
		SchoolID:  schoolPtr(schoolID),
		UserID:    &user.ID,
		Action:    "ARCHIVE",
		TableName: "teachers",
		RecordID:  id.String(),
		NewValues: map[string]any{"is_archived": true},
	})

	c.JSON(http.StatusOK, gin.H{"message": "teacher archived"})
}
