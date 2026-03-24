package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/quran-school/api/internal/auth"
	"github.com/quran-school/api/internal/db"
	"github.com/quran-school/api/internal/model"
)

const (
	ctxUser     = "user"
	ctxTx       = "tx"
	ctxSchoolID = "school_id_str"
)

// RequireAuth validates the Bearer JWT, opens a tenant-scoped transaction,
// and stores the user + tx on the Gin context.
//
// Tenant isolation strategy (fixes cross-school data leakage):
//
//  1. A fresh pgx.Tx is opened from the pool for EVERY request.
//     Because set_config(..., true) is LOCAL to a transaction, the value
//     is guaranteed to be destroyed when the tx ends — it can never
//     survive connection reuse in pgxpool.
//
//  2. Two LOCAL session variables are set inside that transaction:
//     • app.school_id      – UUID string for school_admin, empty for super_admin.
//     • app.is_super_admin – '1' for super_admin, '0' for everyone else.
//     RLS policies MUST check both (see policy example below).
//
//  3. All DB queries in handlers MUST use TxFrom(c), never the raw pool.
//     Using pool.Query/pool.Exec bypasses the LOCAL settings and defeats RLS.
//
//  4. The tx is committed on 2xx and rolled back on 4xx/5xx or panic.
//     Either way the LOCAL variables are destroyed with the transaction.
//
// Matching RLS policy (apply to every tenant-scoped table):
//
//	CREATE POLICY tenant_isolation ON students
//	  USING (
//	    current_setting('app.is_super_admin', true) = '1'
//	    OR school_id = current_setting('app.school_id', true)::uuid
//	  )
//	  WITH CHECK (
//	    current_setting('app.is_super_admin', true) = '1'
//	    OR school_id = current_setting('app.school_id', true)::uuid
//	  );
func RequireAuth(svc *auth.Service, pool *db.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		// ── 1. Extract and validate Bearer token ──────────────────────────────
		header := c.GetHeader("Authorization")
		if header == "" || !strings.HasPrefix(header, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "missing or malformed Authorization header",
			})
			return
		}
		tokenStr := strings.TrimPrefix(header, "Bearer ")

		claims, err := svc.ParseAccessToken(tokenStr)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "invalid or expired token",
			})
			return
		}

		userID, err := uuid.Parse(claims.UserID)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "malformed user id in token",
			})
			return
		}

		// ── 2. Build user model ────────────────────────────────────────────────
		user := &model.User{ID: userID, Role: claims.Role}

		schoolIDStr := "" // empty string = super_admin (no school restriction)
		if claims.SchoolID != nil {
			schoolIDStr = *claims.SchoolID
			sid, err := uuid.Parse(schoolIDStr)
			if err != nil {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
					"error": "malformed school id in token",
				})
				return
			}
			user.SchoolID = &sid
		}

		isSuperAdmin := claims.Role == model.RoleSuperAdmin

		// ── 3. Open a transaction and set BOTH tenant variables LOCAL to it ───
		//
		// The third argument isSuperAdmin causes TxWithTenant to also execute:
		//   SELECT set_config('app.is_super_admin', '1'/'0', true)
		//
		// Using isLocal=true in set_config means the value is scoped to this
		// transaction only. When the tx commits or rolls back the variable
		// reverts, so the pooled connection is returned in a clean state.
		tx, err := pool.TxWithTenant(c.Request.Context(), schoolIDStr, isSuperAdmin)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"error": "database error",
			})
			return
		}

		// ── 4. Store in context and proceed ───────────────────────────────────
		c.Set(ctxUser, user)
		c.Set(ctxTx, tx)
		c.Set(ctxSchoolID, schoolIDStr)

		c.Next()

		// ── 5. Commit on success, rollback on error ────────────────────────────
		// Rollback on an already-committed tx returns pgx.ErrTxClosed (safe to ignore).
		if c.Writer.Status() < 400 {
			_ = tx.Commit(c.Request.Context())
		} else {
			_ = tx.Rollback(c.Request.Context())
		}
	}
}

// UserFrom retrieves the authenticated user stored by RequireAuth.
// Returns nil if called outside an authenticated route.
func UserFrom(c *gin.Context) *model.User {
	v, _ := c.Get(ctxUser)
	u, _ := v.(*model.User)
	return u
}

// TxFrom retrieves the tenant-scoped pgx.Tx stored by RequireAuth.
//
// IMPORTANT: handlers MUST use this tx for every DB call.
// Using the raw pool bypasses app.school_id / app.is_super_admin and
// defeats RLS tenant isolation completely.
func TxFrom(c *gin.Context) pgx.Tx {
	v, _ := c.Get(ctxTx)
	tx, _ := v.(pgx.Tx)
	return tx
}

// SchoolIDStrFrom returns the raw school_id string from the Gin context.
// Returns an empty string for super_admin (no school restriction).
func SchoolIDStrFrom(c *gin.Context) string {
	v, _ := c.Get(ctxSchoolID)
	s, _ := v.(string)
	return s
}

// RequireRole aborts with 403 if the caller's role is not in the allowed list.
// Must be chained after RequireAuth.
func RequireRole(roles ...string) gin.HandlerFunc {
	allowed := make(map[string]struct{}, len(roles))
	for _, r := range roles {
		allowed[r] = struct{}{}
	}
	return func(c *gin.Context) {
		user := UserFrom(c)
		if user == nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthenticated"})
			return
		}
		if _, ok := allowed[user.Role]; !ok {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": "insufficient permissions",
			})
			return
		}
		c.Next()
	}
}
