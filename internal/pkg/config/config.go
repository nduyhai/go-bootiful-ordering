package config

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

// Config represents the application configuration
type Config struct {
	Service   ServiceConfig   `yaml:"service"`
	Jaeger    TempoConfig     `yaml:"jaeger"` // Still using "jaeger" in YAML for backward compatibility
	Tempo     TempoConfig     `yaml:"tempo"`  // New field for explicit Tempo config
	Pyroscope PyroscopeConfig `yaml:"pyroscope"`
	Redis     RedisConfig     `yaml:"redis"`
	DB        DBConfig        `yaml:"db"`
	Server    ServerConfig    `yaml:"server"`
}

// ServiceConfig holds service-specific configuration
type ServiceConfig struct {
	Name string `yaml:"name"`
}

// TempoConfig holds tracing configuration for Tempo
// This replaces JaegerConfig but maintains the same structure
type TempoConfig struct {
	Host     string `yaml:"host"`
	Port     string `yaml:"port"`
	LogSpans bool   `yaml:"logSpans"`
}

// HostPort returns the host:port string for the tracing backend
func (c *TempoConfig) HostPort() string {
	return fmt.Sprintf("%s:%s", c.Host, c.Port)
}

// PyroscopeConfig holds profiling configuration for Pyroscope
type PyroscopeConfig struct {
	Host string `yaml:"host"`
	Port string `yaml:"port"`
}

// ServerAddress returns the server address for the Pyroscope connection
func (c *PyroscopeConfig) ServerAddress() string {
	return fmt.Sprintf("http://%s:%s", c.Host, c.Port)
}

// RedisConfig holds Redis configuration
type RedisConfig struct {
	Host     string `yaml:"host"`
	Port     string `yaml:"port"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`
}

// Addr returns the address for the Redis connection
func (c *RedisConfig) Addr() string {
	return fmt.Sprintf("%s:%s", c.Host, c.Port)
}

// DBConfig holds database configuration
type DBConfig struct {
	Host     string `yaml:"host"`
	Port     string `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Name     string `yaml:"name"`
	SSLMode  string `yaml:"sslMode"`

	// Connection pool settings
	MaxIdleConns    int           `yaml:"maxIdleConns"`
	MaxOpenConns    int           `yaml:"maxOpenConns"`
	ConnMaxLifetime time.Duration `yaml:"connMaxLifetime"`

	// Additional PostgreSQL parameters
	ApplicationName string `yaml:"applicationName"`
	ConnectTimeout  int    `yaml:"connectTimeout"` // in seconds
}

// ServerConfig holds HTTP and gRPC server configuration
type ServerConfig struct {
	HTTP HTTPConfig `yaml:"http"`
	GRPC GRPCConfig `yaml:"grpc"`
}

// HTTPConfig holds HTTP server configuration
type HTTPConfig struct {
	Port string `yaml:"port"`
}

// GRPCConfig holds gRPC server configuration
type GRPCConfig struct {
	Port string `yaml:"port"`
}

// DSN returns the data source name for the database connection in key=value format
func (c *DBConfig) DSN() string {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.Name, c.SSLMode,
	)

	// Add optional parameters if they are set
	if c.ApplicationName != "" {
		dsn += fmt.Sprintf(" application_name=%s", c.ApplicationName)
	}

	if c.ConnectTimeout > 0 {
		dsn += fmt.Sprintf(" connect_timeout=%d", c.ConnectTimeout)
	}

	return dsn
}

// DSNURL returns the data source name for the database connection in URL format
func (c *DBConfig) DSNURL() string {
	// Base URL with credentials and host
	url := fmt.Sprintf("postgres://%s:%s@%s:%s/%s",
		c.User, c.Password, c.Host, c.Port, c.Name)

	// Start query parameters
	params := []string{fmt.Sprintf("sslmode=%s", c.SSLMode)}

	// Add optional parameters if they are set
	if c.ApplicationName != "" {
		params = append(params, fmt.Sprintf("application_name=%s", c.ApplicationName))
	}

	if c.ConnectTimeout > 0 {
		params = append(params, fmt.Sprintf("connect_timeout=%d", c.ConnectTimeout))
	}

	// Join all parameters with &
	if len(params) > 0 {
		url += "?" + params[0]
		for i := 1; i < len(params); i++ {
			url += "&" + params[i]
		}
	}

	return url
}

// Validate checks if the database configuration is valid
func (c *DBConfig) Validate() error {
	if c.Host == "" {
		return fmt.Errorf("database host is required")
	}

	if c.Port == "" {
		return fmt.Errorf("database port is required")
	}

	if c.User == "" {
		return fmt.Errorf("database user is required")
	}

	if c.Name == "" {
		return fmt.Errorf("database name is required")
	}

	return nil
}

// LoadConfig loads configuration from a YAML file
func LoadConfig(configPath string) (*Config, error) {
	// Set default config path if not provided
	if configPath == "" {
		configPath = "config.yaml"
	}

	// Check if the file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("config file not found: %s", configPath)
	}

	// Read the config file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("error reading config file: %w", err)
	}

	// Parse the YAML
	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("error parsing config file: %w", err)
	}

	return &config, nil
}

// LoadServiceConfig loads configuration for a specific service
func LoadServiceConfig(serviceName string) (*Config, error) {
	// Look for config in the current directory
	configPath := fmt.Sprintf("config/%s.yaml", serviceName)

	// If not found, try the parent directory
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		configPath = filepath.Join("..", "config", fmt.Sprintf("%s.yaml", serviceName))
	}

	// If still not found, try the default config
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		configPath = "config/config.yaml"
	}

	// If still not found, try the parent directory default config
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		configPath = filepath.Join("..", "config", "config.yaml")
	}

	return LoadConfig(configPath)
}
