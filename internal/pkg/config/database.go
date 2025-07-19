package config

import (
	"fmt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"log"
	"os"
	"strconv"
	"time"
)

// NewDefaultDBConfig creates a new DBConfig with default values
func NewDefaultDBConfig(dbName string) *DBConfig {
	// Parse connection pool settings from environment variables
	maxIdleConns := 10
	maxOpenConns := 100
	connMaxLifetime := time.Hour

	// Try to parse environment variables for connection pool settings
	if envMaxIdle := getEnv("DB_MAX_IDLE_CONNS", ""); envMaxIdle != "" {
		if val, err := strconv.Atoi(envMaxIdle); err == nil && val > 0 {
			maxIdleConns = val
		}
	}

	if envMaxOpen := getEnv("DB_MAX_OPEN_CONNS", ""); envMaxOpen != "" {
		if val, err := strconv.Atoi(envMaxOpen); err == nil && val > 0 {
			maxOpenConns = val
		}
	}

	if envMaxLifetime := getEnv("DB_CONN_MAX_LIFETIME", ""); envMaxLifetime != "" {
		if val, err := time.ParseDuration(envMaxLifetime); err == nil && val > 0 {
			connMaxLifetime = val
		}
	}

	// Parse connect timeout from environment variable
	connectTimeout := 10 // Default 10 seconds
	if envTimeout := getEnv("DB_CONNECT_TIMEOUT", ""); envTimeout != "" {
		if val, err := strconv.Atoi(envTimeout); err == nil && val > 0 {
			connectTimeout = val
		}
	}

	return &DBConfig{
		// Basic connection parameters
		Host:     getEnv("DB_HOST", "localhost"),
		Port:     getEnv("DB_PORT", "5432"),
		User:     getEnv("DB_USER", "myuser"),
		Password: getEnv("DB_PASSWORD", "secret"),
		Name:     getEnv("DB_NAME", dbName),
		SSLMode:  getEnv("DB_SSL_MODE", "disable"),

		// Connection pool settings
		MaxIdleConns:    maxIdleConns,
		MaxOpenConns:    maxOpenConns,
		ConnMaxLifetime: connMaxLifetime,

		// Additional PostgreSQL parameters
		ApplicationName: getEnv("DB_APPLICATION_NAME", "go-bootiful-ordering"),
		ConnectTimeout:  connectTimeout,
	}
}

// NewGormDB creates a new GORM DB instance from a DBConfig
func NewGormDB(config *DBConfig) (*gorm.DB, error) {
	// Validate the configuration
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid database configuration: %w", err)
	}

	// Configure GORM logger
	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold:             time.Second,
			LogLevel:                  logger.Info,
			IgnoreRecordNotFoundError: true,
			Colorful:                  true,
		},
	)

	// Open connection
	db, err := gorm.Open(postgres.Open(config.DSN()), &gorm.Config{
		Logger: newLogger,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Set connection pool settings from config
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database connection: %w", err)
	}

	// Use connection pool settings from config, or defaults if not set
	maxIdleConns := 10
	if config.MaxIdleConns > 0 {
		maxIdleConns = config.MaxIdleConns
	}

	maxOpenConns := 100
	if config.MaxOpenConns > 0 {
		maxOpenConns = config.MaxOpenConns
	}

	connMaxLifetime := time.Hour
	if config.ConnMaxLifetime > 0 {
		connMaxLifetime = config.ConnMaxLifetime
	}

	sqlDB.SetMaxIdleConns(maxIdleConns)
	sqlDB.SetMaxOpenConns(maxOpenConns)
	sqlDB.SetConnMaxLifetime(connMaxLifetime)

	return db, nil
}

// Helper function to get environment variable with a default value
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
