package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// Config holds all application configuration loaded from environment variables.
type Config struct {
	Port             string
	DatabaseURL      string
	JWTSecret        string
	JWTRefreshSecret string
	AccessTokenTTL   int
	RefreshTokenTTL  int
	Env              string
	MigrationsDir    string
	RunMigrations    bool // RUN_MIGRATIONS=true to auto-apply on startup
}

// Load reads configuration from .env (if present) then from the environment.
func Load() (*Config, error) {
	_ = godotenv.Load()

	cfg := &Config{
		Port:             getEnv("PORT", "8080"),
		DatabaseURL:      mustEnv("DB_URL"),
		JWTSecret:        mustEnv("JWT_SECRET"),
		JWTRefreshSecret: mustEnv("JWT_REFRESH_SECRET"),
		AccessTokenTTL:   getEnvInt("JWT_ACCESS_TTL_MIN", 15),
		RefreshTokenTTL:  getEnvInt("JWT_REFRESH_TTL_DAYS", 7),
		Env:              getEnv("ENV", "development"),
		MigrationsDir:    getEnv("MIGRATIONS_DIR", "./migrations"),
		RunMigrations:    os.Getenv("RUN_MIGRATIONS") == "true",
	}
	return cfg, nil
}

func mustEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		panic(fmt.Sprintf("required environment variable %q is not set", key))
	}
	return v
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return fallback
}
