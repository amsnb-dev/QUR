package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/quran-school/api/internal/model"
)

// canWrite returns true for roles allowed to mutate data.
func canWrite(role string) bool {
	return role == model.RoleSuperAdmin || role == model.RoleSchoolAdmin
}

// RequireWrite aborts with 403 unless the caller is school_admin or super_admin.
// Must be chained after RequireAuth.
func RequireWrite() gin.HandlerFunc {
	return func(c *gin.Context) {
		user := UserFrom(c)
		if user == nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthenticated"})
			return
		}
		if !canWrite(user.Role) {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": "write access requires school_admin or super_admin role",
			})
			return
		}
		c.Next()
	}
}

// RequireNotAccountant aborts with 403 for accountant role.
// Accountants have no access to students/groups.
func RequireNotAccountant() gin.HandlerFunc {
	return func(c *gin.Context) {
		user := UserFrom(c)
		if user == nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthenticated"})
			return
		}
		if user.Role == model.RoleAccountant {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": "accountants do not have access to students or groups",
			})
			return
		}
		c.Next()
	}
}

// RequireSameSchool aborts with 403 if a non-super_admin tries to access
// a resource that belongs to a different school than their own.
//
// Usage: pass the target school UUID extracted from the URL/body.
// Super admins bypass this check entirely (they rely on RLS via app.is_super_admin).
func RequireSameSchool(targetSchoolID uuid.UUID) gin.HandlerFunc {
	return func(c *gin.Context) {
		user := UserFrom(c)
		if user == nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthenticated"})
			return
		}
		// Super admins can cross school boundaries (RLS still applies in DB layer).
		if user.Role == model.RoleSuperAdmin {
			c.Next()
			return
		}
		if user.SchoolID == nil || *user.SchoolID != targetSchoolID {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": "access denied: resource belongs to a different school",
			})
			return
		}
		c.Next()
	}
}

// SchoolIDFrom returns the school UUID from the Gin context (set by RequireAuth).
//
// Returns (uuid.Nil, false) for super_admin — callers must handle this case.
// For all other roles, the UUID is guaranteed to be valid when (ok == true).
func SchoolIDFrom(c *gin.Context) (uuid.UUID, bool) {
	user := UserFrom(c)
	if user == nil || user.SchoolID == nil {
		return uuid.Nil, false
	}
	return *user.SchoolID, true
}
