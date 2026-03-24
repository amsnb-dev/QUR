package handler

import (
"encoding/json"
"net/http"

"github.com/gin-gonic/gin"
"github.com/google/uuid"
"github.com/quran-school/api/internal/middleware"
"github.com/quran-school/api/internal/model"
"github.com/quran-school/api/internal/store"
)

type SettingsHandler struct{}

func NewSettingsHandler() *SettingsHandler { return &SettingsHandler{} }

func (h *SettingsHandler) GetSettings(c *gin.Context) {
tx := middleware.TxFrom(c)
schoolID, _ := middleware.SchoolIDFrom(c)
s, err := store.GetSettings(c.Request.Context(), tx, schoolID)
if err != nil { c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()}); return }
c.JSON(http.StatusOK, s)
}

func (h *SettingsHandler) UpdateSettings(c *gin.Context) {
tx := middleware.TxFrom(c)
schoolID, _ := middleware.SchoolIDFrom(c)
var req model.UpdateSettingsRequest
if err := c.ShouldBindJSON(&req); err != nil { c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()}); return }
s, err := store.UpsertSettings(c.Request.Context(), tx, schoolID, &req)
if err != nil { c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()}); return }
c.JSON(http.StatusOK, s)
}

func (h *SettingsHandler) ListRoles(c *gin.Context) {
tx := middleware.TxFrom(c)
roles, err := store.ListRoles(c.Request.Context(), tx)
if err != nil { c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()}); return }
if roles == nil { roles = []model.SchoolRole{} }
c.JSON(http.StatusOK, gin.H{"data": roles})
}

func (h *SettingsHandler) UpdateRolePermissions(c *gin.Context) {
tx := middleware.TxFrom(c)
id, err := uuid.Parse(c.Param("id"))
if err != nil { c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"}); return }
var perms json.RawMessage
if err := c.ShouldBindJSON(&perms); err != nil { c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()}); return }
r, err := store.UpdateRolePermissions(c.Request.Context(), tx, id, perms)
if err != nil { c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()}); return }
c.JSON(http.StatusOK, r)
}

func (h *SettingsHandler) ListStaff(c *gin.Context) {
tx := middleware.TxFrom(c)
list, err := store.ListStaff(c.Request.Context(), tx)
if err != nil { c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()}); return }
if list == nil { list = []model.SchoolStaff{} }
c.JSON(http.StatusOK, gin.H{"data": list})
}

func (h *SettingsHandler) CreateStaff(c *gin.Context) {
tx := middleware.TxFrom(c)
schoolID, _ := middleware.SchoolIDFrom(c)
var req model.CreateStaffRequest
if err := c.ShouldBindJSON(&req); err != nil { c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()}); return }
s, err := store.CreateStaff(c.Request.Context(), tx, schoolID, &req)
if err != nil { c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()}); return }
c.JSON(http.StatusCreated, s)
}

func (h *SettingsHandler) UpdateStaff(c *gin.Context) {
tx := middleware.TxFrom(c)
id, err := uuid.Parse(c.Param("id"))
if err != nil { c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"}); return }
var req model.UpdateStaffRequest
if err := c.ShouldBindJSON(&req); err != nil { c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()}); return }
s, err := store.UpdateStaff(c.Request.Context(), tx, id, &req)
if err != nil { c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()}); return }
c.JSON(http.StatusOK, s)
}
