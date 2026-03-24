package handler

import (
"net/http"

"github.com/gin-gonic/gin"
"github.com/google/uuid"
"github.com/quran-school/api/internal/middleware"
"github.com/quran-school/api/internal/model"
"github.com/quran-school/api/internal/store"
)

type GuardianHandler struct{}

func NewGuardianHandler() *GuardianHandler { return &GuardianHandler{} }

func (h *GuardianHandler) List(c *gin.Context) {
tx := middleware.TxFrom(c)
f := model.ListGuardiansFilter{Search: c.Query("search")}
list, err := store.ListGuardians(c.Request.Context(), tx, f)
if err != nil {
c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
return
}
if list == nil { list = []model.Guardian{} }
c.JSON(http.StatusOK, gin.H{"data": list, "total": len(list)})
}

func (h *GuardianHandler) Get(c *gin.Context) {
tx := middleware.TxFrom(c)
id, err := uuid.Parse(c.Param("id"))
if err != nil { c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"}); return }
g, err := store.GetGuardian(c.Request.Context(), tx, id)
if err != nil { c.JSON(http.StatusNotFound, gin.H{"error": "not found"}); return }
c.JSON(http.StatusOK, g)
}

func (h *GuardianHandler) Create(c *gin.Context) {
tx := middleware.TxFrom(c)
schoolID, _ := middleware.SchoolIDFrom(c)
var req model.CreateGuardianRequest
if err := c.ShouldBindJSON(&req); err != nil {
c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
return
}
g, err := store.CreateGuardian(c.Request.Context(), tx, schoolID, &req)
if err != nil { c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()}); return }
c.JSON(http.StatusCreated, g)
}

func (h *GuardianHandler) Update(c *gin.Context) {
tx := middleware.TxFrom(c)
id, err := uuid.Parse(c.Param("id"))
if err != nil { c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"}); return }
var req model.UpdateGuardianRequest
if err := c.ShouldBindJSON(&req); err != nil {
c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
return
}
g, err := store.UpdateGuardian(c.Request.Context(), tx, id, &req)
if err != nil { c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()}); return }
c.JSON(http.StatusOK, g)
}

func (h *GuardianHandler) GetStudents(c *gin.Context) {
tx := middleware.TxFrom(c)
id, err := uuid.Parse(c.Param("id"))
if err != nil { c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"}); return }
rows, err := tx.Query(c.Request.Context(), `
SELECT id, full_name, status, level_on_entry, memorized_parts
FROM students WHERE guardian_id = $1 AND is_archived = false
ORDER BY full_name
`, id)
if err != nil { c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()}); return }
defer rows.Close()
type StudentSummary struct {
ID            string  `json:"id"`
FullName      string  `json:"full_name"`
Status        string  `json:"status"`
LevelOnEntry  *string `json:"level_on_entry"`
MemorizedParts float64 `json:"memorized_parts"`
}
var students []StudentSummary
for rows.Next() {
var s StudentSummary
if err := rows.Scan(&s.ID, &s.FullName, &s.Status, &s.LevelOnEntry, &s.MemorizedParts); err != nil { continue }
students = append(students, s)
}
if students == nil { students = []StudentSummary{} }
c.JSON(http.StatusOK, gin.H{"data": students})
}
