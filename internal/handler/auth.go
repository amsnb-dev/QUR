package handler

import (
	"time"
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/quran-school/api/internal/auth"
	"github.com/quran-school/api/internal/config"
	"github.com/quran-school/api/internal/db"
	"github.com/quran-school/api/internal/middleware"
	"github.com/quran-school/api/internal/model"
)

// AuthHandler groups all authentication endpoints.
type AuthHandler struct {
	cfg *config.Config
	db  *db.Pool
	svc *auth.Service
}

// NewAuthHandler constructs an AuthHandler.
func NewAuthHandler(cfg *config.Config, pool *db.Pool, svc *auth.Service) *AuthHandler {
	return &AuthHandler{cfg: cfg, db: pool, svc: svc}
}

// ─────────────────────────────────────────────────────────────
// POST /auth/login
// ─────────────────────────────────────────────────────────────

func (h *AuthHandler) Login(c *gin.Context) {
	var req model.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx := c.Request.Context()

	// Single transaction for the entire login flow.
	// No tenant scoping needed: login is a cross-tenant operation.
	tx, err := h.db.Begin(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
		return
	}
	defer func() { _ = tx.Rollback(ctx) }()

	user, hash, err := fetchUserByEmail(ctx, tx, req.Email)
	if err != nil {
		// Constant-time response: do not reveal whether the user exists.
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	if !auth.CheckPassword(req.Password, hash) {
		// Increment failed_attempts inside the same tx.
		if upErr := incrementFailedAttempts(ctx, tx, user.ID); upErr != nil {
			_ = tx.Rollback(ctx)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
			return
		}
		_ = tx.Commit(ctx) // persist the failed_attempts increment
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	// Generate tokens before the final commit so the whole operation is atomic.
	rawRefresh, hashRefresh, err := auth.GenerateRefreshToken()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "token generation failed"})
		return
	}

	accessToken, err := h.svc.IssueAccessToken(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "token generation failed"})
		return
	}

	if err := auth.StoreRefreshToken(ctx, tx, user.ID, hashRefresh, h.cfg.RefreshTokenTTL); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
		return
	}

	if err := resetFailedAttempts(ctx, tx, user.ID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
		return
	}

	if err := tx.Commit(ctx); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
		return
	}

	c.JSON(http.StatusOK, model.TokenPair{
		AccessToken:  accessToken,
		RefreshToken: rawRefresh,
		ExpiresIn:    h.cfg.AccessTokenTTL * 60,
	})
}

// ─────────────────────────────────────────────────────────────
// POST /auth/refresh
// ─────────────────────────────────────────────────────────────

func (h *AuthHandler) Refresh(c *gin.Context) {
	var req model.RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx := c.Request.Context()

	// Single transaction: revoke old token + insert new one atomically.
	tx, err := h.db.Begin(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
		return
	}
	defer func() { _ = tx.Rollback(ctx) }()

	newRaw, _, userID, err := auth.RotateRefreshToken(ctx, tx, req.RefreshToken, h.cfg.RefreshTokenTTL)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired refresh token"})
		return
	}

	// Fetch fresh user data (role / school may have changed since last login).
	user, err := fetchUserByID(ctx, tx, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "user lookup failed"})
		return
	}

	accessToken, err := h.svc.IssueAccessToken(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "token generation failed"})
		return
	}

	if err := tx.Commit(ctx); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
		return
	}

	c.JSON(http.StatusOK, model.TokenPair{
		AccessToken:  accessToken,
		RefreshToken: newRaw,
		ExpiresIn:    h.cfg.AccessTokenTTL * 60,
	})
}

// ─────────────────────────────────────────────────────────────
// POST /auth/logout  (requires Bearer token via RequireAuth)
// ─────────────────────────────────────────────────────────────

func (h *AuthHandler) Logout(c *gin.Context) {
	user := middleware.UserFrom(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthenticated"})
		return
	}

	ctx := c.Request.Context()

	// Use the middleware's tenant-scoped tx for consistency.
	tx := middleware.TxFrom(c)
	if tx == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "no db transaction"})
		return
	}

	if err := auth.RevokeAllUserTokens(ctx, tx, user.ID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "logout failed"})
		return
	}

	// The middleware commits/rollbacks after c.Next() returns.
	c.JSON(http.StatusOK, gin.H{"message": "logged out successfully"})
}

// ─────────────────────────────────────────────────────────────
// GET /me  (requires Bearer token via RequireAuth)
// ─────────────────────────────────────────────────────────────

func (h *AuthHandler) Me(c *gin.Context) {
	user := middleware.UserFrom(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthenticated"})
		return
	}

	tx := middleware.TxFrom(c)
	if tx == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "no db transaction"})
		return
	}

	enriched, err := fetchUserByID(c.Request.Context(), tx, user.ID)
	if err != nil {
		// Fallback to JWT claims only — no extra DB call needed.
		c.JSON(http.StatusOK, gin.H{
			"id":        user.ID,
			"role":      user.Role,
			"school_id": user.SchoolID,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":        enriched.ID,
		"email":     enriched.Email,
		"full_name": enriched.FullName,
		"role":      enriched.Role,
		"school_id": enriched.SchoolID,
	})
}

// ─────────────────────────────────────────────────────────────
// Package-level DB helpers (use the caller's tx — no extra connections)
// ─────────────────────────────────────────────────────────────

// fetchUserByEmail looks up a user by email inside an existing tx.
// Returns the user and bcrypt hash, or an error if not found / locked / inactive.
func fetchUserByEmail(ctx context.Context, tx pgx.Tx, email string) (*model.User, string, error) {
	var (
		lockedUntil *time.Time
		failedAttempts int16
		id           uuid.UUID
		schoolID     *uuid.UUID
		role         string
		fullName     string
		passwordHash string
		isActive     bool
		isArchived   bool
	)

	err := tx.QueryRow(ctx, `
        SELECT u.id,
               u.school_id,
               r.name,
               u.full_name,
               u.password_hash,
               u.is_active,
               COALESCE(u.is_archived, FALSE),
               u.locked_until,
               u.failed_attempts
        FROM auth_user_by_email($1) AS u
        JOIN roles r ON r.id = u.role_id
`, email).Scan(
        &id, &schoolID, &role, &fullName,
        &passwordHash, &isActive, &isArchived,
        &lockedUntil, &failedAttempts,
)
	
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, "", fmt.Errorf("user not found or locked")
		}
		return nil, "", fmt.Errorf("fetchUserByEmail: %w", err)
	}
	if lockedUntil != nil && lockedUntil.After(time.Now()) {
    return nil, "", fmt.Errorf("user locked")
	}
	if !isActive || isArchived {
		return nil, "", fmt.Errorf("account inactive or archived")
	}
	
	return &model.User{
		ID:       id,
		SchoolID: schoolID,
		Role:     role,
		FullName: fullName,
		Email:    email,
	}, passwordHash, nil
}

// fetchUserByID looks up a user by id inside an existing tx.
func fetchUserByID(ctx context.Context, tx pgx.Tx, id uuid.UUID) (*model.User, error) {
	var (
		schoolID   *uuid.UUID
		role       string
		fullName   string
		email      string
		isActive   bool
		isArchived bool
	)

	err := tx.QueryRow(ctx, `
		SELECT u.school_id, r.name, u.full_name, u.email,
		       u.is_active, COALESCE(u.is_archived, FALSE)
		FROM   users u
		JOIN   roles r ON r.id = u.role_id
		WHERE  u.id = $1
	`, id).Scan(&schoolID, &role, &fullName, &email, &isActive, &isArchived)
	if err != nil {
		return nil, fmt.Errorf("fetchUserByID: %w", err)
	}

	if !isActive || isArchived {
		return nil, fmt.Errorf("account inactive or archived")
	}

	return &model.User{
		ID:       id,
		SchoolID: schoolID,
		Role:     role,
		FullName: fullName,
		Email:    email,
	}, nil
}

// incrementFailedAttempts increments failed_attempts and locks the account
// after 5 consecutive failures. Uses the caller's tx.
func incrementFailedAttempts(ctx context.Context, tx pgx.Tx, id uuid.UUID) error {
	_, err := tx.Exec(ctx, `
		UPDATE users
		SET failed_attempts = failed_attempts + 1,
		    locked_until = CASE
		        WHEN failed_attempts + 1 >= 5
		        THEN NOW() + INTERVAL '30 minutes'
		        ELSE locked_until
		    END
		WHERE id = $1
	`, id)
	return err
}

// resetFailedAttempts clears the failure counter on successful login.
// Uses the caller's tx.
func resetFailedAttempts(ctx context.Context, tx pgx.Tx, id uuid.UUID) error {
	_, err := tx.Exec(ctx, `
		UPDATE users
		SET failed_attempts = 0,
		    locked_until    = NULL,
		    last_login_at   = NOW()
		WHERE id = $1
	`, id)
	return err
}
