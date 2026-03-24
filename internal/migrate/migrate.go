// Package migrate runs ordered *.sql files from a directory against the DB.
// It keeps a simple migration_log table to track which files have run.
// Safe to call on every startup — already-applied files are skipped.
package migrate

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Run applies every *.sql file in dir that has not yet been recorded
// in the migration_log table. Files are applied in lexicographic order.
func Run(ctx context.Context, pool *pgxpool.Pool, dir string) error {
	// Ensure the tracking table exists (idempotent).
	if _, err := pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS migration_log (
			filename   TEXT        PRIMARY KEY,
			applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)
	`); err != nil {
		return fmt.Errorf("migrate: create log table: %w", err)
	}

	// Load list of already-applied files.
	rows, err := pool.Query(ctx, "SELECT filename FROM migration_log ORDER BY filename")
	if err != nil {
		return fmt.Errorf("migrate: query log: %w", err)
	}
	applied := make(map[string]bool)
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return err
		}
		applied[name] = true
	}
	rows.Close()

	// Collect *.sql files in order.
	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("migrate: read dir %q: %w", dir, err)
	}

	var files []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".sql") {
			files = append(files, e.Name())
		}
	}
	sort.Strings(files)

	// Apply pending files.
	for _, name := range files {
		if applied[name] {
			log.Printf("migrate: skip %s (already applied)", name)
			continue
		}

		content, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			return fmt.Errorf("migrate: read %s: %w", name, err)
		}

		tx, err := pool.Begin(ctx)
		if err != nil {
			return fmt.Errorf("migrate: begin tx for %s: %w", name, err)
		}

		if _, err := tx.Exec(ctx, string(content)); err != nil {
			_ = tx.Rollback(ctx)
			return fmt.Errorf("migrate: apply %s: %w", name, err)
		}

		if _, err := tx.Exec(ctx,
			"INSERT INTO migration_log (filename) VALUES ($1)", name,
		); err != nil {
			_ = tx.Rollback(ctx)
			return fmt.Errorf("migrate: record %s: %w", name, err)
		}

		if err := tx.Commit(ctx); err != nil {
			return fmt.Errorf("migrate: commit %s: %w", name, err)
		}

		log.Printf("migrate: applied %s ✅", name)
	}

	return nil
}
