package db

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"io"
	"io/fs"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"path"
	"sort"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

// DB wraps sql.DB with additional methods
type DB struct {
	*sql.DB
}

// Connect opens a connection to SQLite database
func Connect(dataSourceName string) (*DB, error) {
	// Ensure the directory for the database file exists
	dir := filepath.Dir(dataSourceName)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create database directory: %w", err)
		}
	}

	// Ensure the database file exists on disk to avoid lazy creation issues in packaged builds
	if _, err := os.Stat(dataSourceName); os.IsNotExist(err) {
		f, createErr := os.OpenFile(dataSourceName, os.O_RDWR|os.O_CREATE, 0644)
		if createErr != nil {
			return nil, fmt.Errorf("failed to create database file: %w", createErr)
		}
		_ = f.Close()
	}

	// Build a SQLite DSN that explicitly enables read/write/create mode
	absPath, _ := filepath.Abs(dataSourceName)
	// Convert Windows backslashes to forward slashes for the URI
	uriPath := strings.ReplaceAll(absPath, "\\", "/")
	// Construct DSN without over-escaping (sqlite accepts file:C:/... on Windows)
	dsn := fmt.Sprintf("file:%s?mode=rwc&cache=shared", uriPath)

	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Basic sanity ping early (with timeout)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("database ping failed: %w (dsn=%s)", err, dsn)
	}

	// Enable foreign keys
	if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("failed to enable foreign keys: %w", err)
	}

	// Improve concurrency and reduce locking issues
	if _, err := db.Exec("PRAGMA journal_mode=WAL"); err != nil {
		// Non-fatal: log via fmt, but continue
		log.Printf("[db] warning: failed to set journal_mode=WAL: %v", err)
	}
	if _, err := db.Exec("PRAGMA busy_timeout=5000"); err != nil {
		log.Printf("[db] warning: failed to set busy_timeout: %v", err)
	}

	// Set connection pool settings
	db.SetMaxOpenConns(10)   // Allow multiple connections for better concurrency
	db.SetMaxIdleConns(5)    // Keep some idle connections
	db.SetConnMaxLifetime(0) // No connection lifetime limit

	wrapped := &DB{DB: db}

	// Optional deep diagnostics when DEBUG_DB=1
	if v, ok := os.LookupEnv("DEBUG_DB"); ok && v != "" && v != "0" && strings.ToLower(v) != "false" {
		log.Printf("[db] DEBUG_DB enabled. dsn=%s", dsn)
		// SQLite version
		var sqliteVersion string
		if err := wrapped.QueryRow("select sqlite_version()").Scan(&sqliteVersion); err == nil {
			log.Printf("[db] sqlite_version=%s", sqliteVersion)
		}
		// List attached databases
		rows, err := wrapped.Query("PRAGMA database_list")
		if err == nil {
			defer rows.Close()
			for rows.Next() {
				var seq int
				var name, file string
				if scanErr := rows.Scan(&seq, &name, &file); scanErr == nil {
					log.Printf("[db] database_list seq=%d name=%s file=%s", seq, name, file)
				}
			}
		}
		// List tables
		tableRows, err := wrapped.Query("SELECT name FROM sqlite_master WHERE type='table' ORDER BY name")
		if err == nil {
			defer tableRows.Close()
			var tables []string
			for tableRows.Next() {
				var t string
				if scanErr := tableRows.Scan(&t); scanErr == nil {
					tables = append(tables, t)
				}
			}
			if len(tables) > 0 {
				log.Printf("[db] existing tables: %s", strings.Join(tables, ", "))
			} else {
				log.Printf("[db] no tables present yet (migrations pending)")
			}
		}
	}

	return wrapped, nil
}

// RunMigrations executes all migration files in the migrations directory
func (db *DB) RunMigrations(migrationsDir string) error {
	// Create migrations table if it doesn't exist
	createMigrationsTable := `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version TEXT PRIMARY KEY,
			applied_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		);
	`
	if _, err := db.Exec(createMigrationsTable); err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	// Get all applied migrations
	appliedMigrations := make(map[string]bool)
	rows, err := db.Query("SELECT version FROM schema_migrations")
	if err != nil {
		return fmt.Errorf("failed to query applied migrations: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var version string
		if err := rows.Scan(&version); err != nil {
			return fmt.Errorf("failed to scan migration version: %w", err)
		}
		appliedMigrations[version] = true
	}

	// Get all migration files
	migrationFiles, err := filepath.Glob(filepath.Join(migrationsDir, "*.up.sql"))
	if err != nil {
		return fmt.Errorf("failed to glob migration files: %w", err)
	}

	// Sort migration files
	sort.Strings(migrationFiles)

	// Apply pending migrations
	for _, migrationFile := range migrationFiles {
		// Extract version from filename (e.g., "001_initial_schema.up.sql" -> "001_initial_schema")
		filename := filepath.Base(migrationFile)
		version := strings.TrimSuffix(filename, ".up.sql")

		if appliedMigrations[version] {
			continue // Migration already applied
		}

		// Read migration file
		content, err := ioutil.ReadFile(migrationFile)
		if err != nil {
			return fmt.Errorf("failed to read migration file %s: %w", migrationFile, err)
		}

		// Execute migration
		if _, err := db.Exec(string(content)); err != nil {
			return fmt.Errorf("failed to execute migration %s: %w", version, err)
		}

		// Record migration as applied
		if _, err := db.Exec("INSERT INTO schema_migrations (version) VALUES (?)", version); err != nil {
			return fmt.Errorf("failed to record migration %s: %w", version, err)
		}

		fmt.Printf("Applied migration: %s\n", version)
	}

	return nil
}

// RunEmbeddedMigrations executes all migration files from embedded filesystem
func (db *DB) RunEmbeddedMigrations(embeddedFS embed.FS, migrationsDir string) error {
	// Create migrations table if it doesn't exist
	createMigrationsTable := `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version TEXT PRIMARY KEY,
			applied_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		);
	`
	if _, err := db.Exec(createMigrationsTable); err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	// Get all applied migrations
	appliedMigrations := make(map[string]bool)
	rows, err := db.Query("SELECT version FROM schema_migrations")
	if err != nil {
		return fmt.Errorf("failed to query applied migrations: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var version string
		if err := rows.Scan(&version); err != nil {
			return fmt.Errorf("failed to scan migration version: %w", err)
		}
		appliedMigrations[version] = true
	}

	// Get all migration files from embedded filesystem
	entries, err := fs.ReadDir(embeddedFS, migrationsDir)
	if err != nil {
		return fmt.Errorf("failed to read embedded migrations directory: %w", err)
	}

	var migrationFiles []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".up.sql") {
			migrationFiles = append(migrationFiles, entry.Name())
		}
	}

	// Sort migration files
	sort.Strings(migrationFiles)

	// Apply pending migrations
	for _, migrationFile := range migrationFiles {
		// Extract version from filename (e.g., "001_initial_schema.up.sql" -> "001_initial_schema")
		version := strings.TrimSuffix(migrationFile, ".up.sql")

		if appliedMigrations[version] {
			continue // Migration already applied
		}

		// Read migration file from embedded filesystem
	// Use forward-slash join for embedded FS paths
	fullPath := path.Join(migrationsDir, migrationFile)
		content, err := embeddedFS.ReadFile(fullPath)
		if err != nil {
			return fmt.Errorf("failed to read embedded migration file %s: %w", migrationFile, err)
		}

		// Execute migration
		if _, err := db.Exec(string(content)); err != nil {
			return fmt.Errorf("failed to execute migration %s: %w", version, err)
		}

		// Record migration as applied
		if _, err := db.Exec("INSERT INTO schema_migrations (version) VALUES (?)", version); err != nil {
			return fmt.Errorf("failed to record migration %s: %w", version, err)
		}

		fmt.Printf("Applied migration: %s\n", version)
	}

	return nil
}

// ApplyEmbeddedSchema reads a single consolidated SQL schema from an embedded FS and applies it idempotently.
// The schema should contain only CREATE IF NOT EXISTS / ALTER TABLE ADD COLUMN IF NOT EXISTS / DROP VIEW IF EXISTS etc.
func (db *DB) ApplyEmbeddedSchema(fsys embed.FS, schemaPath string) error {
	// Read schema SQL
	content, err := fsys.ReadFile(schemaPath)
	if err != nil {
		return fmt.Errorf("failed to read embedded schema %s: %w", schemaPath, err)
	}
	return db.applySQLBatch(string(content))
}

// ApplySchemaFile applies a schema from a disk file (useful in dev).
func (db *DB) ApplySchemaFile(path string) error {
	f, err := os.Open(path)
	if err != nil { return fmt.Errorf("open schema file: %w", err) }
	defer f.Close()
	data, err := io.ReadAll(f)
	if err != nil { return fmt.Errorf("read schema file: %w", err) }
	return db.applySQLBatch(string(data))
}

// applySQLBatch executes multiple SQL statements separated by semicolons.
// It ignores empty statements and trims whitespace. Errors abort processing.
func (db *DB) applySQLBatch(sqlBatch string) error {
	// Remove block comments /* ... */
	cleaned := removeBlockComments(sqlBatch)
	// Remove single-line comments starting with --
	var b strings.Builder
	for _, line := range strings.Split(cleaned, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "--") { continue }
		// Also strip inline -- comments if present
		if idx := strings.Index(trimmed, "--"); idx >= 0 {
			trimmed = strings.TrimSpace(trimmed[:idx])
		}
		if trimmed != "" {
			b.WriteString(trimmed)
			b.WriteString("\n")
		}
	}

	// Execute within a transaction to ensure atomicity where possible.
	// Some DDL in SQLite causes implicit commits; we accept partial application if statements succeed individually.
	// So we run statements individually for resilience.
	stmts := strings.Split(b.String(), ";")
	for _, raw := range stmts {
		stmt := strings.TrimSpace(raw)
		if stmt == "" { continue }
		// Handle broader compatibility for older SQLite versions that don't support
		// "ALTER TABLE ... ADD COLUMN IF NOT EXISTS ..."
		if strings.Contains(strings.ToUpper(stmt), "ALTER TABLE") &&
		   strings.Contains(strings.ToUpper(stmt), "ADD COLUMN IF NOT EXISTS") {
			// Try executing as-is first (newer SQLite)
			if _, err := db.Exec(stmt); err != nil {
				// Fallback: strip IF NOT EXISTS and retry; ignore duplicate column errors
				fallback := strings.Replace(strings.Replace(stmt, " IF NOT EXISTS", "", 1), " if not exists", "", 1)
				if _, err2 := db.Exec(fallback); err2 != nil {
					msg := strings.ToLower(err2.Error())
					if strings.Contains(msg, "duplicate column") || strings.Contains(msg, "already exists") {
						// Safe to ignore: column already present
						continue
					}
					return fmt.Errorf("schema apply error on statement: %s\n%w", fallback, err2)
				}
			}
			continue
		}

		if _, err := db.Exec(stmt); err != nil {
			// For idempotent creates, some drivers may not support IF NOT EXISTS everywhere; be lenient on 'already exists'
			msg := strings.ToLower(err.Error())
			if strings.Contains(msg, "already exists") || strings.Contains(msg, "duplicate column") {
				continue
			}
			return fmt.Errorf("schema apply error on statement: %s\n%w", stmt, err)
		}
	}
	return nil
}

// removeBlockComments removes /* ... */ style comments from SQL text.
func removeBlockComments(s string) string {
	for {
		start := strings.Index(s, "/*")
		if start < 0 { break }
		end := strings.Index(s[start+2:], "*/")
		if end < 0 {
			// Unclosed comment; drop rest
			s = s[:start]
			break
		}
		end += start + 2
		s = s[:start] + s[end+2:]
	}
	return s
}
