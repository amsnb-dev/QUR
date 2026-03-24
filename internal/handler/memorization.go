package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/quran-school/api/internal/middleware"
	"github.com/quran-school/api/internal/model"
	"github.com/quran-school/api/internal/store"
)

// MemorizationHandler handles memorization endpoints.
type MemorizationHandler struct{}

func NewMemorizationHandler() *MemorizationHandler { return &MemorizationHandler{} }

// POST /students/:id/memorization
func (h *MemorizationHandler) Create(c *gin.Context) {
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

	var req model.CreateMemorizationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Default entry_type to "new"
	if req.EntryType == "" {
		req.EntryType = "new"
	}

	// Teacher write-access: must be assigned to a group the student belongs to.
	if user.Role == "teacher" {
		if req.GroupID == nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "teachers must supply group_id"})
			return
		}
		teacherUserID, err := store.GetGroupTeacherUserID(c.Request.Context(), tx, *req.GroupID)
		if err != nil {
			if err == store.ErrNotFound {
				c.JSON(http.StatusNotFound, gin.H{"error": "group not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if teacherUserID != user.ID {
			c.JSON(http.StatusForbidden, gin.H{"error": "teachers may only record memorization for their own groups"})
			return
		}
	}

	rec, err := store.CreateMemorization(c.Request.Context(), tx, schoolID, studentID, user.ID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Audit
	nv := map[string]any{
		"surah_number": rec.SurahNumber,
		"from_verse":   rec.FromVerse,
		"to_verse":     rec.ToVerse,
		"entry_type":   rec.EntryType,
		"grade":        rec.Grade,
	}
	_ = store.InsertAudit(c.Request.Context(), tx, store.AuditParams{
		SchoolID:  &schoolID,
		UserID:    &user.ID,
		Action:    "INSERT",
		TableName: "memorization_logs",
		RecordID:  rec.ID.String(),
		NewValues: nv,
	})

	c.JSON(http.StatusCreated, rec)
}

// GET /students/:id/memorization
func (h *MemorizationHandler) List(c *gin.Context) {
	tx := middleware.TxFrom(c)
	studentID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid student id"})
		return
	}

	p := store.ParsePage(c.Query("limit"), c.Query("offset"))

	records, total, err := store.ListMemorization(c.Request.Context(), tx, studentID, p)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if records == nil {
		records = []model.MemorizationRecord{}
	}
	c.JSON(http.StatusOK, store.PagedResult[model.MemorizationRecord]{
		Data: records,
		Meta: store.Meta{Limit: p.Limit, Offset: p.Offset, Total: total},
	})
}

// PATCH /memorization/:id
func (h *MemorizationHandler) Update(c *gin.Context) {
	tx := middleware.TxFrom(c)
	user := middleware.UserFrom(c)

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid memorization id"})
		return
	}

	// Fetch for RBAC + school verification.
	existingSchoolID, existingStudentID, err := store.GetMemorizationSchoolAndStudent(c.Request.Context(), tx, id)
	if err != nil {
		if err == store.ErrNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "memorization record not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Verify tenant ownership (RLS already filters, but double-check for non-super).
	schoolID, ok := resolveSchoolID(c)
	if ok && schoolID != existingSchoolID {
		c.JSON(http.StatusNotFound, gin.H{"error": "memorization record not found"})
		return
	}

	// Teacher: must be assigned to a group the student belongs to.
	if user.Role == "teacher" {
		allowed, err := store.TeacherCanAccessStudent(c.Request.Context(), tx, user.ID, existingStudentID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if !allowed {
			c.JSON(http.StatusForbidden, gin.H{"error": "teachers may only edit memorization for students in their groups"})
			return
		}
	}

	var req model.UpdateMemorizationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	rec, err := store.UpdateMemorization(c.Request.Context(), tx, id, &req)
	if err != nil {
		if err == store.ErrNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "memorization record not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Audit
	nv := map[string]any{}
	if req.Grade != nil {
		nv["grade"] = *req.Grade
	}
	if req.Notes != nil {
		nv["notes"] = *req.Notes
	}
	if req.FromVerse != nil {
		nv["from_verse"] = *req.FromVerse
	}
	if req.ToVerse != nil {
		nv["to_verse"] = *req.ToVerse
	}
	_ = store.InsertAudit(c.Request.Context(), tx, store.AuditParams{
		SchoolID:  &existingSchoolID,
		UserID:    &user.ID,
		Action:    "UPDATE",
		TableName: "memorization_logs",
		RecordID:  id.String(),
		NewValues: nv,
	})

	c.JSON(http.StatusOK, rec)
}
