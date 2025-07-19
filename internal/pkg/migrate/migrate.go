package migrate

import (
	"fmt"
	"log"
	"path/filepath"
	"strings"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

// Config holds the configuration for migrations
type Config struct {
	Enabled bool          // Whether migrations are enabled
	Service string        // Service name (order or product)
	DSN     string        // Database connection string
	Dir     string        // Migration directory (optional)
	Timeout time.Duration // Timeout for migrations
}

// NewDefaultConfig creates a new Config with default values
func NewDefaultConfig(service string, dsn string) *Config {
	return &Config{
		Enabled: true,
		Service: service,
		DSN:     dsn,
		Dir:     "",
		Timeout: 30 * time.Second,
	}
}

// Run runs database migrations
func Run(cfg *Config) error {
	if !cfg.Enabled {
		log.Println("Migrations are disabled")
		return nil
	}

	// Validate service
	if cfg.Service == "" {
		return fmt.Errorf("service is required")
	}
	if cfg.Service != "order" && cfg.Service != "product" {
		return fmt.Errorf("invalid service: %s. Must be 'order' or 'product'", cfg.Service)
	}

	// Set migration directory
	migrationDir := cfg.Dir
	if migrationDir == "" {
		// Use default directory with SQL subdirectory for golang-migrate
		migrationDir = filepath.Join("migrations", cfg.Service, "sql")
	}

	log.Printf("Running migrations for service %s from directory %s", cfg.Service, migrationDir)

	// Run migrations with timeout
	return RunWithTimeout(migrationDir, cfg.DSN, cfg.Timeout)
}

// RunWithTimeout runs migrations with a timeout
func RunWithTimeout(dir, dsn string, timeout time.Duration) error {
	done := make(chan error, 1)

	go func() {
		done <- runMigrations(dir, dsn)
	}()

	select {
	case err := <-done:
		return err
	case <-time.After(timeout):
		return fmt.Errorf("migration timed out after %s", timeout)
	}
}

// runMigrations runs database migrations using golang-migrate
func runMigrations(dir, dsn string) error {
	// Convert backslashes to forward slashes for URL compatibility
	dirWithForwardSlashes := strings.ReplaceAll(dir, "\\", "/")

	// Create source URL for migrations
	sourceURL := fmt.Sprintf("file://%s", dirWithForwardSlashes)

	// Create a new migrate instance
	m, err := migrate.New(sourceURL, dsn)
	if err != nil {
		return fmt.Errorf("failed to create migrate instance: %w", err)
	}
	defer m.Close()

	// Set logger
	m.Log = &MigrateLogger{}

	// Run migrations
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to apply migrations: %w", err)
	}

	return nil
}

// MigrateLogger implements migrate.Logger interface
type MigrateLogger struct{}

// Printf logs a formatted string
func (l *MigrateLogger) Printf(format string, v ...interface{}) {
	log.Printf(format, v...)
}

// Verbose returns whether verbose output is enabled
func (l *MigrateLogger) Verbose() bool {
	return true
}
