package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/quran-school/api/internal/config"
	"github.com/quran-school/api/internal/model"
	"golang.org/x/crypto/bcrypt"
)

const bcryptCost = 12

// Service handles token issuance and validation.
type Service struct{ cfg *config.Config }

func NewService(cfg *config.Config) *Service { return &Service{cfg: cfg} }

// ── Password ──────────────────────────────────────────────────

// HashPassword returns a bcrypt hash (cost 12).
func HashPassword(password string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCost)
	if err != nil {
		return "", fmt.Errorf("hash password: %w", err)
	}
	return string(b), nil
}

// CheckPassword compares plaintext against a stored bcrypt hash.
func CheckPassword(password, hash string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}

// ── JWT ───────────────────────────────────────────────────────

type jwtClaims struct {
	jwt.RegisteredClaims
	UserID   string  `json:"uid"`
	SchoolID *string `json:"sid"`
	Role     string  `json:"role"`
}

// IssueAccessToken signs a JWT containing uid, sid, role.
func (s *Service) IssueAccessToken(user *model.User) (string, error) {
	now := time.Now()
	ttl := time.Duration(s.cfg.AccessTokenTTL) * time.Minute

	var sid *string
	if user.SchoolID != nil {
		v := user.SchoolID.String()
		sid = &v
	}

	claims := jwtClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   user.ID.String(),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
		},
		UserID:   user.ID.String(),
		SchoolID: sid,
		Role:     user.Role,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(s.cfg.JWTSecret))
	if err != nil {
		return "", fmt.Errorf("sign token: %w", err)
	}
	return signed, nil
}

// ParseAccessToken validates signature + expiry and returns the embedded claims.
func (s *Service) ParseAccessToken(tokenStr string) (*model.AccessClaims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &jwtClaims{},
		func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
			}
			return []byte(s.cfg.JWTSecret), nil
		},
		jwt.WithExpirationRequired(),
	)
	if err != nil {
		return nil, fmt.Errorf("parse token: %w", err)
	}

	c, ok := token.Claims.(*jwtClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token claims")
	}

	return &model.AccessClaims{
		UserID:   c.UserID,
		SchoolID: c.SchoolID,
		Role:     c.Role,
	}, nil
}

// ── Refresh Tokens ────────────────────────────────────────────

// GenerateRefreshToken returns (rawToken, sha256Hex).
// Only rawToken is sent to the client; only sha256Hex is stored.
func GenerateRefreshToken() (raw, hashed string, err error) {
	b := make([]byte, 32)
	if _, err = rand.Read(b); err != nil {
		return "", "", fmt.Errorf("generate token: %w", err)
	}
	raw = hex.EncodeToString(b)
	hashed = hashToken(raw)
	return raw, hashed, nil
}

func hashToken(raw string) string {
	h := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(h[:])
}

// StoreRefreshToken inserts a hashed token row within an existing tx.
func StoreRefreshToken(ctx context.Context, tx pgx.Tx, userID uuid.UUID, hash string, ttlDays int) error {
	expiresAt := time.Now().Add(time.Duration(ttlDays) * 24 * time.Hour)
	_, err := tx.Exec(ctx, `
		INSERT INTO refresh_tokens (user_id, token_hash, expires_at)
		VALUES ($1, $2, $3)
	`, userID, hash, expiresAt)
	if err != nil {
		return fmt.Errorf("store refresh token: %w", err)
	}
	return nil
}

// RotateRefreshToken validates the raw token, revokes it, inserts a new one.
// Returns (newRaw, newHash, userID).
func RotateRefreshToken(
	ctx context.Context, tx pgx.Tx, rawToken string, ttlDays int,
) (newRaw, newHash string, userID uuid.UUID, err error) {
	hash := hashToken(rawToken)

	var (
		rtID      uuid.UUID
		uid       uuid.UUID
		expiresAt time.Time
		revoked   bool
	)

	err = tx.QueryRow(ctx, `
		SELECT id, user_id, expires_at, revoked
		FROM   refresh_tokens
		WHERE  token_hash = $1
	`, hash).Scan(&rtID, &uid, &expiresAt, &revoked)
	if err != nil {
		if err == pgx.ErrNoRows {
			return "", "", uuid.Nil, fmt.Errorf("refresh token not found")
		}
		return "", "", uuid.Nil, fmt.Errorf("lookup token: %w", err)
	}

	if revoked {
		return "", "", uuid.Nil, fmt.Errorf("refresh token already revoked")
	}
	if time.Now().After(expiresAt) {
		return "", "", uuid.Nil, fmt.Errorf("refresh token expired")
	}

	// Revoke old (token rotation — one-time use)
	if _, err = tx.Exec(ctx,
		"UPDATE refresh_tokens SET revoked = TRUE WHERE id = $1", rtID,
	); err != nil {
		return "", "", uuid.Nil, fmt.Errorf("revoke old token: %w", err)
	}

	newRaw, newHash, err = GenerateRefreshToken()
	if err != nil {
		return "", "", uuid.Nil, err
	}
	if err = StoreRefreshToken(ctx, tx, uid, newHash, ttlDays); err != nil {
		return "", "", uuid.Nil, err
	}

	return newRaw, newHash, uid, nil
}

// RevokeAllUserTokens revokes every active refresh token for a user (logout).
func RevokeAllUserTokens(ctx context.Context, tx pgx.Tx, userID uuid.UUID) error {
	_, err := tx.Exec(ctx, `
		UPDATE refresh_tokens
		SET    revoked = TRUE
		WHERE  user_id = $1
		  AND  revoked = FALSE
	`, userID)
	if err != nil {
		return fmt.Errorf("revoke all tokens: %w", err)
	}
	return nil
}
