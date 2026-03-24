package auth_test

import (
	"testing"

	"github.com/quran-school/api/internal/auth"
)

func TestHashAndCheckPassword(t *testing.T) {
	password := "Admin1234!"

	hash, err := auth.HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword error: %v", err)
	}

	if !auth.CheckPassword(password, hash) {
		t.Fatal("CheckPassword: expected true for correct password")
	}

	if auth.CheckPassword("WrongPassword!", hash) {
		t.Fatal("CheckPassword: expected false for wrong password")
	}
}

func TestGenerateRefreshToken(t *testing.T) {
	raw1, hash1, err := auth.GenerateRefreshToken()
	if err != nil {
		t.Fatalf("GenerateRefreshToken error: %v", err)
	}

	raw2, hash2, err := auth.GenerateRefreshToken()
	if err != nil {
		t.Fatalf("GenerateRefreshToken error: %v", err)
	}

	// Two tokens must be different
	if raw1 == raw2 {
		t.Fatal("expected unique raw tokens")
	}
	if hash1 == hash2 {
		t.Fatal("expected unique hashes")
	}

	// Hash must not equal raw
	if raw1 == hash1 {
		t.Fatal("hash must not equal raw token")
	}

	// Length: 32 random bytes → 64 hex chars
	if len(raw1) != 64 {
		t.Fatalf("expected 64-char raw token, got %d", len(raw1))
	}
	if len(hash1) != 64 {
		t.Fatalf("expected 64-char hash, got %d", len(hash1))
	}
}

func TestBcryptCost(t *testing.T) {
	// Verify bcrypt cost is 12 by checking hash prefix
	hash, err := auth.HashPassword("testpassword")
	if err != nil {
		t.Fatalf("HashPassword error: %v", err)
	}
	// bcrypt hashes with cost 12 start with "$2a$12$"
	if len(hash) < 7 || hash[:7] != "$2a$12$" {
		t.Fatalf("expected bcrypt cost 12 hash prefix '$2a$12$', got prefix: %q", hash[:7])
	}
}
