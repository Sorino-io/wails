package db

import (
	"database/sql"
	"embed"
	"fmt"
	"io/fs"
	"io/ioutil"
	"path/filepath"
	"sort"
	"strings"

	_ "modernc.org/sqlite"
)

// DB wraps sql.DB with additional methods
type DB struct {
	*sql.DB
}

// Connect opens a connection to SQLite database
func Connect(dataSourceName string) (*DB, error) {
	db, err := sql.Open("sqlite", dataSourceName)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Enable foreign keys
	if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		return nil, fmt.Errorf("failed to enable foreign keys: %w", err)
	}

	// Set connection pool settings
	db.SetMaxOpenConns(10)   // Allow multiple connections for better concurrency
	db.SetMaxIdleConns(5)    // Keep some idle connections
	db.SetConnMaxLifetime(0) // No connection lifetime limit

	return &DB{DB: db}, nil
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
		fullPath := filepath.Join(migrationsDir, migrationFile)
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
