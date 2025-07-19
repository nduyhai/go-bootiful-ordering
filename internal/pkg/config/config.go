package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Config represents the application configuration
type Config struct {
	Service   ServiceConfig   `yaml:"service" mapstructure:"service"`
	Jaeger    TempoConfig     `yaml:"jaeger" mapstructure:"jaeger"` // Still using "jaeger" in YAML for backward compatibility
	Tempo     TempoConfig     `yaml:"tempo" mapstructure:"tempo"`   // New field for explicit Tempo config
	Pyroscope PyroscopeConfig `yaml:"pyroscope" mapstructure:"pyroscope"`
	Redis     RedisConfig     `yaml:"redis" mapstructure:"redis"`
	DB        DBConfig        `yaml:"db" mapstructure:"db"`
	Server    ServerConfig    `yaml:"server" mapstructure:"server"`
}

// ServiceConfig holds service-specific configuration
type ServiceConfig struct {
	Name string `yaml:"name" mapstructure:"name"`
}

// TempoConfig holds tracing configuration for Tempo
// This replaces JaegerConfig but maintains the same structure
type TempoConfig struct {
	Host     string `yaml:"host" mapstructure:"host"`
	Port     string `yaml:"port" mapstructure:"port"`
	LogSpans bool   `yaml:"logSpans" mapstructure:"logSpans"`
}

// HostPort returns the host:port string for the tracing backend
func (c *TempoConfig) HostPort() string {
	return fmt.Sprintf("%s:%s", c.Host, c.Port)
}

// PyroscopeConfig holds profiling configuration for Pyroscope
type PyroscopeConfig struct {
	Host string `yaml:"host" mapstructure:"host"`
	Port string `yaml:"port" mapstructure:"port"`
}

// ServerAddress returns the server address for the Pyroscope connection
func (c *PyroscopeConfig) ServerAddress() string {
	return fmt.Sprintf("http://%s:%s", c.Host, c.Port)
}

// RedisConfig holds Redis configuration
type RedisConfig struct {
	Host     string `yaml:"host" mapstructure:"host"`
	Port     string `yaml:"port" mapstructure:"port"`
	Password string `yaml:"password" mapstructure:"password"`
	DB       int    `yaml:"db" mapstructure:"db"`
}

// Addr returns the address for the Redis connection
func (c *RedisConfig) Addr() string {
	return fmt.Sprintf("%s:%s", c.Host, c.Port)
}

// DBConfig holds database configuration
type DBConfig struct {
	Host     string `yaml:"host" mapstructure:"host"`
	Port     string `yaml:"port" mapstructure:"port"`
	User     string `yaml:"user" mapstructure:"user"`
	Password string `yaml:"password" mapstructure:"password"`
	Name     string `yaml:"name" mapstructure:"name"`
	SSLMode  string `yaml:"sslMode" mapstructure:"sslMode"`

	// Connection pool settings
	MaxIdleConns    int           `yaml:"maxIdleConns" mapstructure:"maxIdleConns"`
	MaxOpenConns    int           `yaml:"maxOpenConns" mapstructure:"maxOpenConns"`
	ConnMaxLifetime time.Duration `yaml:"connMaxLifetime" mapstructure:"connMaxLifetime"`

	// Additional PostgreSQL parameters
	ApplicationName string `yaml:"applicationName" mapstructure:"applicationName"`
	ConnectTimeout  int    `yaml:"connectTimeout" mapstructure:"connectTimeout"` // in seconds
}

// ServerConfig holds HTTP and gRPC server configuration
type ServerConfig struct {
	HTTP HTTPConfig `yaml:"http" mapstructure:"http"`
	GRPC GRPCConfig `yaml:"grpc" mapstructure:"grpc"`
}

// HTTPConfig holds HTTP server configuration
type HTTPConfig struct {
	Port string `yaml:"port" mapstructure:"port"`
}

// GRPCConfig holds gRPC server configuration
type GRPCConfig struct {
	Port string `yaml:"port" mapstructure:"port"`
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

// LoadConfig loads configuration using Viper
func LoadConfig(configPath string) (*Config, error) {
	v := viper.New()

	// Set default config path if not provided
	if configPath == "" {
		configPath = "config.yaml"
	}

	// Extract the directory and filename from the path
	configDir := filepath.Dir(configPath)
	configName := strings.TrimSuffix(filepath.Base(configPath), filepath.Ext(configPath))

	// Configure Viper to read from the config file
	v.SetConfigName(configName)
	v.SetConfigType("yaml")
	v.AddConfigPath(configDir)

	// Check if the file exists before attempting to read it
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("config file not found: %s", configPath)
	}

	// Configure Viper to read from environment variables
	v.SetEnvPrefix("")                                 // No prefix for environment variables
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_")) // Replace dots with underscores in env vars
	v.AutomaticEnv()                                   // Read environment variables that match

	// Read the config file
	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("error reading config file: %w", err)
	}

	// Unmarshal the config into our struct
	var config Config
	if err := v.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("error parsing config: %w", err)
	}

	return &config, nil
}

// LoadServiceConfig loads configuration for a specific service using Viper
func LoadServiceConfig(serviceName string) (*Config, error) {
	v := viper.New()

	// Configure Viper to read from environment variables first
	v.SetEnvPrefix("")                                 // No prefix for environment variables
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_")) // Replace dots with underscores in env vars
	v.AutomaticEnv()                                   // Read environment variables that match

	// Try to find the config file in different locations
	configPaths := []string{
		fmt.Sprintf("config/%s.yaml", serviceName),
		filepath.Join("..", "config", fmt.Sprintf("%s.yaml", serviceName)),
		"config/config.yaml",
		filepath.Join("..", "config", "config.yaml"),
	}

	// Set the config name and type
	v.SetConfigType("yaml")

	// Try each config path
	var configFound bool
	for _, path := range configPaths {
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			configDir := filepath.Dir(path)
			configName := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))

			v.SetConfigName(configName)
			v.AddConfigPath(configDir)
			configFound = true
			break
		}
	}

	// If no config file was found, return an error
	if !configFound {
		return nil, fmt.Errorf("no config file found for service: %s", serviceName)
	}

	// Read the config file
	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("error reading config file: %w", err)
	}

	// Unmarshal the config into our struct
	var config Config
	if err := v.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("error parsing config: %w", err)
	}

	return &config, nil
}
