package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/quran-school/api/internal/middleware"
	"github.com/quran-school/api/internal/model"
	"github.com/quran-school/api/internal/store"
)

type UserHandler struct{}

func NewUserHandler() *UserHandler { return &UserHandler{} }

// GET /users
func (h *UserHandler) List(c *gin.Context) {
	tx := middleware.TxFrom(c)
	f := model.ListUsersFilter{
		IncludeArchived: c.Query("include_archived") == "1",
		RoleName:        c.Query("role"),
	}
	p := store.ParsePage(c.Query("limit"), c.Query("offset"))

	users, total, err := store.ListUsers(c.Request.Context(), tx, f, p)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if users == nil {
		users = []model.UserRecord{}
	}
	c.JSON(http.StatusOK, store.PagedResult[model.UserRecord]{
		Data: users,
		Meta: store.Meta{Limit: p.Limit, Offset: p.Offset, Total: total},
	})
}

// GET /users/:id
func (h *UserHandler) Get(c *gin.Context) {
	tx := middleware.TxFrom(c)
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}
	u, err := store.GetUser(c.Request.Context(), tx, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if u == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}
	c.JSON(http.StatusOK, u)
}

// POST /users
func (h *UserHandler) Create(c *gin.Context) {
	tx := middleware.TxFrom(c)
	caller := middleware.UserFrom(c)

	var req model.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Determine school: super_admin can specify, school_admin uses own school
	var schoolID *uuid.UUID
	if caller.Role == model.RoleSuperAdmin {
		if req.SchoolID != nil {
			sid, err := uuid.Parse(*req.SchoolID)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid school_id"})
				return
			}
			schoolID = &sid
		}
	} else {
		schoolID = caller.SchoolID
	}

	u, err := store.CreateUser(c.Request.Context(), tx, schoolID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	_ = store.InsertAudit(c.Request.Context(), tx, store.AuditParams{
		SchoolID:  schoolID,
		UserID:    &caller.ID,
		Action:    "INSERT",
		TableName: "users",
		RecordID:  u.ID.String(),
		NewValues: map[string]any{"full_name": u.FullName, "email": u.Email, "role": u.RoleName},
	})

	c.JSON(http.StatusCreated, u)
}

// PUT /users/:id
func (h *UserHandler) Update(c *gin.Context) {
	tx := middleware.TxFrom(c)
	caller := middleware.UserFrom(c)

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	var req model.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	u, err := store.UpdateUser(c.Request.Context(), tx, id, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if u == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	_ = store.InsertAudit(c.Request.Context(), tx, store.AuditParams{
		UserID:    &caller.ID,
		Action:    "UPDATE",
		TableName: "users",
		RecordID:  id.String(),
		NewValues: map[string]any{"full_name": u.FullName, "is_active": u.IsActive},
	})

	c.JSON(http.StatusOK, u)
}

// PATCH /users/:id/archive
func (h *UserHandler) Archive(c *gin.Context) {
	tx := middleware.TxFrom(c)
	caller := middleware.UserFrom(c)

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	// Prevent self-archive
	if id == caller.ID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot archive your own account"})
		return
	}

	if err := store.ArchiveUser(c.Request.Context(), tx, id); err != nil {
		if err == store.ErrNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found or already inactive"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	_ = store.InsertAudit(c.Request.Context(), tx, store.AuditParams{
		UserID:    &caller.ID,
		Action:    "ARCHIVE",
		TableName: "users",
		RecordID:  id.String(),
		NewValues: map[string]any{"is_active": false},
	})

	c.JSON(http.StatusOK, gin.H{"message": "user archived"})
}
