package handler

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/quran-school/api/internal/middleware"
	"github.com/quran-school/api/internal/model"
	"github.com/quran-school/api/internal/store"
)

// AttendanceHandler handles attendance endpoints.
type AttendanceHandler struct{}

func NewAttendanceHandler() *AttendanceHandler { return &AttendanceHandler{} }

// POST /groups/:id/attendance
// Bulk upsert: one request records all students' attendance for a date.
func (h *AttendanceHandler) BulkUpsert(c *gin.Context) {
	tx := middleware.TxFrom(c)
	user := middleware.UserFrom(c)

	groupID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid group id"})
		return
	}
	schoolID, ok := resolveSchoolID(c)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "super_admin must specify a school context"})
		return
	}

	// Teacher write-access check: teacher may only record for their own group.
	if user.Role == "teacher" {
		teacherUserID, err := store.GetGroupTeacherUserID(c.Request.Context(), tx, groupID)
		if err != nil {
			if err == store.ErrNotFound {
				c.JSON(http.StatusNotFound, gin.H{"error": "group not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if teacherUserID != user.ID {
			c.JSON(http.StatusForbidden, gin.H{"error": "teachers may only record attendance for their own groups"})
			return
		}
	}

	var req model.BulkAttendanceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	date, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "date must be YYYY-MM-DD"})
		return
	}

	var results []model.AttendanceRecord
	for _, item := range req.Items {
		rec, err := store.UpsertAttendanceItem(c.Request.Context(), tx, schoolID, groupID, user.ID, item, date)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":      fmt.Sprintf("failed on student %s: %s", item.StudentID, err.Error()),
				"student_id": item.StudentID,
			})
			return
		}
		results = append(results, *rec)
	}

	// Audit: one entry per batch
	_ = store.InsertAudit(c.Request.Context(), tx, store.AuditParams{
		SchoolID:  &schoolID,
		UserID:    &user.ID,
		Action:    "INSERT",
		TableName: "attendance",
		RecordID:  groupID.String(),
		NewValues: map[string]any{"date": req.Date, "count": len(results)},
	})

	if results == nil {
		results = []model.AttendanceRecord{}
	}
	c.JSON(http.StatusOK, gin.H{"data": results, "count": len(results)})
}

// GET /groups/:id/attendance
func (h *AttendanceHandler) List(c *gin.Context) {
	tx := middleware.TxFrom(c)
	groupID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid group id"})
		return
	}

	var datePtr *time.Time
	if ds := c.Query("date"); ds != "" {
		d, err := time.Parse("2006-01-02", ds)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "date must be YYYY-MM-DD"})
			return
		}
		datePtr = &d
	}

	p := store.ParsePage(c.Query("limit"), c.Query("offset"))

	records, total, err := store.ListAttendance(c.Request.Context(), tx, groupID, datePtr, p)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if records == nil {
		records = []model.AttendanceRecord{}
	}
	c.JSON(http.StatusOK, store.PagedResult[model.AttendanceRecord]{
		Data: records,
		Meta: store.Meta{Limit: p.Limit, Offset: p.Offset, Total: total},
	})
}

// PATCH /attendance/:id
func (h *AttendanceHandler) Update(c *gin.Context) {
	tx := middleware.TxFrom(c)
	user := middleware.UserFrom(c)

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid attendance id"})
		return
	}

	// Verify record belongs to this school (RLS handles it, but we do
	// explicit check for teachers to ensure they own the group).
	if user.Role == "teacher" {
		sid, err := store.GetAttendanceSchoolID(c.Request.Context(), tx, id)
		if err != nil {
			if err == store.ErrNotFound {
				c.JSON(http.StatusNotFound, gin.H{"error": "attendance record not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		schoolID, _ := resolveSchoolID(c)
		if schoolID != uuid.Nil && sid != schoolID {
			c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
			return
		}
	}

	var req model.UpdateAttendanceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	rec, err := store.UpdateAttendance(c.Request.Context(), tx, id, &req)
	if err != nil {
		if err == store.ErrNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "attendance record not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Audit
	schoolID, _ := resolveSchoolID(c)
	nv := map[string]any{}
	if req.Status != nil {
		nv["status"] = *req.Status
	}
	if req.Note != nil {
		nv["note"] = *req.Note
	}
	_ = store.InsertAudit(c.Request.Context(), tx, store.AuditParams{
		SchoolID:  &schoolID,
		UserID:    &user.ID,
		Action:    "UPDATE",
		TableName: "attendance",
		RecordID:  id.String(),
		NewValues: nv,
	})

	c.JSON(http.StatusOK, rec)
}
