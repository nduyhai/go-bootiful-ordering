package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config represents the application configuration
type Config struct {
	Service ServiceConfig `yaml:"service"`
	Jaeger  JaegerConfig  `yaml:"jaeger"`
	Redis   RedisConfig   `yaml:"redis"`
	DB      DBConfig      `yaml:"db"`
	Server  ServerConfig  `yaml:"server"`
}

// ServiceConfig holds service-specific configuration
type ServiceConfig struct {
	Name string `yaml:"name"`
}

// JaegerConfig holds Jaeger tracing configuration
type JaegerConfig struct {
	Host     string `yaml:"host"`
	Port     string `yaml:"port"`
	LogSpans bool   `yaml:"logSpans"`
}

// HostPort returns the host:port string for Jaeger
func (c *JaegerConfig) HostPort() string {
	return fmt.Sprintf("%s:%s", c.Host, c.Port)
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

// DSN returns the data source name for the database connection
func (c *DBConfig) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.Name, c.SSLMode,
	)
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
