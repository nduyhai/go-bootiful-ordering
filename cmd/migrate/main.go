package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func main() {
	// Parse command-line flags
	var (
		service = flag.String("service", "", "Service to migrate (order or product)")
		dsn     = flag.String("dsn", "", "Database connection string (optional, will use environment variables if not provided)")
		dir     = flag.String("dir", "", "Migration directory (optional, will use default if not provided)")
		dryRun  = flag.Bool("dry-run", false, "Dry run (don't apply migrations)")
		devURL  = flag.String("dev-url", "", "Dev database URL for schema diff (optional)")
	)
	flag.Parse()

	// Validate service
	if *service == "" {
		log.Fatal("Service is required. Use -service=order or -service=product")
	}
	if *service != "order" && *service != "product" {
		log.Fatalf("Invalid service: %s. Must be 'order' or 'product'", *service)
	}

	// Set migration directory
	migrationDir := *dir
	if migrationDir == "" {
		// Use default directory
		migrationDir = filepath.Join("migrations", *service)
	}

	// Check if migration directory exists
	if _, err := os.Stat(migrationDir); os.IsNotExist(err) {
		log.Fatalf("Migration directory does not exist: %s", migrationDir)
	}

	// Build DSN if not provided
	if *dsn == "" {
		*dsn = buildDSN(*service)
	}

	// Run Atlas migrate
	if err := runAtlasMigrate(migrationDir, *dsn, *dryRun, *devURL); err != nil {
		log.Fatalf("Migration failed: %v", err)
	}

	log.Printf("Migration completed successfully for service: %s", *service)
}

// buildDSN builds a DSN from environment variables
func buildDSN(service string) string {
	// Use uppercase prefix for environment variables
	prefix := strings.ToUpper(service)

	// Get database configuration from environment variables
	host := getEnv(fmt.Sprintf("%s_DB_HOST", prefix), "localhost")
	port := getEnv(fmt.Sprintf("%s_DB_PORT", prefix), "5432")
	user := getEnv(fmt.Sprintf("%s_DB_USER", prefix), "myuser")
	password := getEnv(fmt.Sprintf("%s_DB_PASSWORD", prefix), "secret")
	dbName := getEnv(fmt.Sprintf("%s_DB_NAME", prefix), service)
	sslMode := getEnv(fmt.Sprintf("%s_DB_SSL_MODE", prefix), "disable")

	// Build DSN
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		user, password, host, port, dbName, sslMode,
	)
}

// runAtlasMigrate runs Atlas migrate command
func runAtlasMigrate(dir, dsn string, dryRun bool, devURL string) error {
	// Build Atlas command
	args := []string{
		"migrate",
		"apply",
		"--dir", fmt.Sprintf("file://%s", dir),
		"--url", dsn,
	}

	// Add dry-run flag if specified
	if dryRun {
		args = append(args, "--dry-run")
	}

	// Add dev-url if specified (for schema diff)
	if devURL != "" {
		args = append(args, "--dev-url", devURL)
	}

	// Create command
	cmd := exec.CommandContext(context.Background(), "atlas", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Run command
	return cmd.Run()
}

// getEnv gets an environment variable with a default value
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
