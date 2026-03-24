package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/quran-school/api/internal/middleware"
	"github.com/quran-school/api/internal/model"
	"github.com/quran-school/api/internal/store"
)

// GroupHandler handles /groups endpoints.
type GroupHandler struct{}

func NewGroupHandler() *GroupHandler { return &GroupHandler{} }

// resolveSchoolID returns the tenant UUID from the authenticated user's JWT claims.
// Returns uuid.Nil, false for super_admin (no school affiliation).
func resolveSchoolID(c *gin.Context) (uuid.UUID, bool) {
	user := middleware.UserFrom(c)
	if user.SchoolID == nil {
		return uuid.Nil, false
	}
	return *user.SchoolID, true
}

// schoolPtr wraps a UUID as a pointer, returning nil for uuid.Nil.
func schoolPtr(id uuid.UUID) *uuid.UUID {
	if id == uuid.Nil {
		return nil
	}
	return &id
}

// POST /groups
func (h *GroupHandler) Create(c *gin.Context) {
	tx := middleware.TxFrom(c)
	user := middleware.UserFrom(c)
	schoolID, ok := resolveSchoolID(c)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "super_admin must specify a school context"})
		return
	}

	var req model.CreateGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	g, err := store.CreateGroup(c.Request.Context(), tx, schoolID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	_ = store.InsertAudit(c.Request.Context(), tx, store.AuditParams{
		SchoolID:  schoolPtr(schoolID),
		UserID:    &user.ID,
		Action:    "INSERT",
		TableName: "groups",
		RecordID:  g.ID.String(),
		NewValues: map[string]any{"name": g.Name, "teacher_id": g.TeacherID},
	})

	c.JSON(http.StatusCreated, g)
}

// GET /groups
func (h *GroupHandler) List(c *gin.Context) {
	tx := middleware.TxFrom(c)
	f := model.ListGroupsFilter{
		IncludeArchived: c.Query("include_archived") == "1",
	}
	p := store.ParsePage(c.Query("limit"), c.Query("offset"))

	groups, total, err := store.ListGroups(c.Request.Context(), tx, f, p)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if groups == nil {
		groups = []model.Group{}
	}
	c.JSON(http.StatusOK, store.PagedResult[model.Group]{
		Data: groups,
		Meta: store.Meta{Limit: p.Limit, Offset: p.Offset, Total: total},
	})
}

// GET /groups/:id
func (h *GroupHandler) Get(c *gin.Context) {
	tx := middleware.TxFrom(c)
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid group id"})
		return
	}
	g, err := store.GetGroup(c.Request.Context(), tx, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if g == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "group not found"})
		return
	}
	c.JSON(http.StatusOK, g)
}

// PUT /groups/:id
func (h *GroupHandler) Update(c *gin.Context) {
	tx := middleware.TxFrom(c)
	user := middleware.UserFrom(c)
	schoolID, _ := resolveSchoolID(c)

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid group id"})
		return
	}

	var req model.UpdateGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	g, err := store.UpdateGroup(c.Request.Context(), tx, id, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if g == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "group not found"})
		return
	}

	_ = store.InsertAudit(c.Request.Context(), tx, store.AuditParams{
		SchoolID:  schoolPtr(schoolID),
		UserID:    &user.ID,
		Action:    "UPDATE",
		TableName: "groups",
		RecordID:  id.String(),
		NewValues: map[string]any{"name": g.Name, "is_active": g.IsActive},
	})

	c.JSON(http.StatusOK, g)
}

// PATCH /groups/:id/archive
func (h *GroupHandler) Archive(c *gin.Context) {
	tx := middleware.TxFrom(c)
	user := middleware.UserFrom(c)
	schoolID, _ := resolveSchoolID(c)

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid group id"})
		return
	}

	if err := store.ArchiveGroup(c.Request.Context(), tx, id, user.ID, schoolPtr(schoolID)); err != nil {
		if err == store.ErrNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "group not found or already archived"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Audit already written inside store.ArchiveGroup via DB function.
	// Add an application-level entry as well for full traceability.
	_ = store.InsertAudit(c.Request.Context(), tx, store.AuditParams{
		SchoolID:  schoolPtr(schoolID),
		UserID:    &user.ID,
		Action:    "ARCHIVE",
		TableName: "groups",
		RecordID:  id.String(),
		NewValues: map[string]any{"is_archived": true},
	})

	c.JSON(http.StatusOK, gin.H{"message": "group archived"})
}
