package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/quran-school/api/internal/middleware"
	"github.com/quran-school/api/internal/model"
	"github.com/quran-school/api/internal/store"
)

// SchoolHandler handles /schools endpoints — super_admin only.
type SchoolHandler struct{}

func NewSchoolHandler() *SchoolHandler { return &SchoolHandler{} }

// GET /schools
func (h *SchoolHandler) List(c *gin.Context) {
	tx := middleware.TxFrom(c)
	schools, err := store.ListSchools(c.Request.Context(), tx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if schools == nil {
		schools = []model.School{}
	}
	c.JSON(http.StatusOK, gin.H{"data": schools, "total": len(schools)})
}

// GET /schools/:id
func (h *SchoolHandler) Get(c *gin.Context) {
	tx := middleware.TxFrom(c)
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid school id"})
		return
	}
	sc, err := store.GetSchool(c.Request.Context(), tx, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if sc == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "school not found"})
		return
	}
	c.JSON(http.StatusOK, sc)
}

// POST /schools
func (h *SchoolHandler) Create(c *gin.Context) {
	tx := middleware.TxFrom(c)
	var req model.CreateSchoolRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	sc, err := store.CreateSchool(c.Request.Context(), tx, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, sc)
}

// PUT /schools/:id
func (h *SchoolHandler) Update(c *gin.Context) {
	tx := middleware.TxFrom(c)
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid school id"})
		return
	}
	var req model.UpdateSchoolRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	sc, err := store.UpdateSchool(c.Request.Context(), tx, id, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if sc == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "school not found"})
		return
	}
	c.JSON(http.StatusOK, sc)
}
