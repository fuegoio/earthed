// Package migrations applies embedded SQL migrations to a database.
package migrations

import (
	"database/sql"
	"embed"
	"fmt"
	"io/fs"
	"sort"
	"strings"
)

//go:embed *.sql
var migrationFS embed.FS

// Run applies any pending embedded SQL migrations to db, in filename order.
func Run(db *sql.DB) error {
	if _, err := db.Exec(`CREATE TABLE IF NOT EXISTS schema_migrations (
		filename VARCHAR(255) PRIMARY KEY,
		applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
	)`); err != nil {
		return fmt.Errorf("create schema_migrations: %w", err)
	}

	files, err := fs.ReadDir(migrationFS, ".")
	if err != nil {
		return fmt.Errorf("read migrations: %w", err)
	}

	sort.Slice(files, func(i, j int) bool { return files[i].Name() < files[j].Name() })

	for _, f := range files {
		if !strings.HasSuffix(f.Name(), ".sql") {
			continue
		}
		var exists bool
		if err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE filename=$1)", f.Name()).Scan(&exists); err != nil {
			return fmt.Errorf("check %s: %w", f.Name(), err)
		}
		if exists {
			continue
		}
		content, err := migrationFS.ReadFile(f.Name())
		if err != nil {
			return fmt.Errorf("read %s: %w", f.Name(), err)
		}
		tx, err := db.Begin()
		if err != nil {
			return fmt.Errorf("begin %s: %w", f.Name(), err)
		}
		if _, err := tx.Exec(string(content)); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("exec %s: %w", f.Name(), err)
		}
		if _, err := tx.Exec("INSERT INTO schema_migrations(filename) VALUES($1)", f.Name()); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("record %s: %w", f.Name(), err)
		}
		if err := tx.Commit(); err != nil {
			return fmt.Errorf("commit %s: %w", f.Name(), err)
		}
	}
	return nil
}
