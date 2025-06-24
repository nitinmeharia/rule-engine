package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Load loads configuration from files and environment variables
func Load() (*Config, error) {
	// Set defaults
	setDefaults()

	// Configure Viper
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./configs")
	viper.AddConfigPath("../configs")
	viper.AddConfigPath("../../configs")
	viper.AddConfigPath(".")

	// Enable environment variable binding
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.SetEnvPrefix("RULE_ENGINE")

	// Read base configuration
	if err := viper.ReadInConfig(); err != nil {
		// Config file is optional
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
	}

	// Get environment from ENV or config
	environment := viper.GetString("environment")
	if environment == "" {
		environment = "development"
	}

	// Load environment-specific config if it exists
	envConfigName := fmt.Sprintf("config.%s", environment)
	viper.SetConfigName(envConfigName)

	if err := viper.MergeInConfig(); err != nil {
		// Environment-specific config is optional
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to merge environment config: %w", err)
		}
	}

	// Unmarshal into config struct
	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Validate configuration
	if err := validateConfig(&config); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return &config, nil
}

// setDefaults sets reasonable default values for configuration
func setDefaults() {
	// Environment
	viper.SetDefault("environment", "development")

	// Server defaults
	viper.SetDefault("server.port", 8080)
	viper.SetDefault("server.host", "0.0.0.0")
	viper.SetDefault("server.readTimeout", "30s")
	viper.SetDefault("server.writeTimeout", "30s")
	viper.SetDefault("server.idleTimeout", "120s")
	viper.SetDefault("server.shutdownTimeout", "30s")
	viper.SetDefault("server.enableCORS", true)
	viper.SetDefault("server.trustedProxies", []string{})

	// Database defaults
	viper.SetDefault("database.host", "localhost")
	viper.SetDefault("database.port", 5432)
	viper.SetDefault("database.name", "rule_engine")
	viper.SetDefault("database.user", "postgres")
	viper.SetDefault("database.password", "")
	viper.SetDefault("database.sslMode", "disable")
	viper.SetDefault("database.maxConnections", 10)
	viper.SetDefault("database.maxIdleConns", 5)
	viper.SetDefault("database.connMaxLifetime", "1h")
	viper.SetDefault("database.queryTimeout", "10s")

	// Cache defaults
	viper.SetDefault("cache.refreshIntervalSec", 60)
	viper.SetDefault("cache.refreshJitterSec", 10)
	viper.SetDefault("cache.refreshTimeout", "30s")
	viper.SetDefault("cache.maxSizeMB", 256)

	// Circuit breaker defaults
	viper.SetDefault("cache.circuitBreaker.enabled", true)
	viper.SetDefault("cache.circuitBreaker.failureThreshold", 5)
	viper.SetDefault("cache.circuitBreaker.recoveryTimeout", "60s")
	viper.SetDefault("cache.circuitBreaker.maxRequests", 3)

	// JWT defaults
	viper.SetDefault("jwt.secret", "change-me-in-production")
	viper.SetDefault("jwt.tokenExpiration", "24h")
	viper.SetDefault("jwt.requiredClaims", []string{"clientId", "role"})
	viper.SetDefault("jwt.skipExpiryCheck", false)

	// Metrics defaults
	viper.SetDefault("metrics.enabled", true)
	viper.SetDefault("metrics.path", "/v1/metrics")
	viper.SetDefault("metrics.port", 0) // Use same port as server
	viper.SetDefault("metrics.namespace", "rule_engine")
	viper.SetDefault("metrics.subsystem", "api")

	// Logger defaults
	viper.SetDefault("logger.level", "info")
	viper.SetDefault("logger.format", "console")
	viper.SetDefault("logger.enableCaller", false)
	viper.SetDefault("logger.timeFormat", time.RFC3339)
}

// validateConfig validates the loaded configuration for required fields and constraints
func validateConfig(config *Config) error {
	// Server validation
	if config.Server.Port < 1 || config.Server.Port > 65535 {
		return fmt.Errorf("server.port must be between 1 and 65535, got %d", config.Server.Port)
	}

	if config.Server.Host == "" {
		return fmt.Errorf("server.host cannot be empty")
	}

	// Database validation
	if config.Database.URL == "" {
		// Validate components if URL not provided
		if config.Database.Host == "" {
			return fmt.Errorf("database.host cannot be empty when database.url is not provided")
		}
		if config.Database.Name == "" {
			return fmt.Errorf("database.name cannot be empty when database.url is not provided")
		}
		if config.Database.User == "" {
			return fmt.Errorf("database.user cannot be empty when database.url is not provided")
		}
	}

	if config.Database.MaxConnections < 1 {
		return fmt.Errorf("database.maxConnections must be at least 1, got %d", config.Database.MaxConnections)
	}

	if config.Database.MaxIdleConns < 0 {
		return fmt.Errorf("database.maxIdleConns cannot be negative, got %d", config.Database.MaxIdleConns)
	}

	if config.Database.MaxIdleConns > config.Database.MaxConnections {
		return fmt.Errorf("database.maxIdleConns (%d) cannot exceed maxConnections (%d)",
			config.Database.MaxIdleConns, config.Database.MaxConnections)
	}

	// Cache validation
	if config.Cache.RefreshIntervalSec < 1 {
		return fmt.Errorf("cache.refreshIntervalSec must be at least 1, got %d", config.Cache.RefreshIntervalSec)
	}

	if config.Cache.RefreshJitterSec < 0 {
		return fmt.Errorf("cache.refreshJitterSec cannot be negative, got %d", config.Cache.RefreshJitterSec)
	}

	if config.Cache.MaxSizeMB < 1 {
		return fmt.Errorf("cache.maxSizeMB must be at least 1, got %d", config.Cache.MaxSizeMB)
	}

	// JWT validation
	if config.JWT.Secret == "" {
		return fmt.Errorf("jwt.secret cannot be empty")
	}

	if config.JWT.Secret == "change-me-in-production" && config.IsProduction() {
		return fmt.Errorf("jwt.secret must be changed in production environment")
	}

	if len(config.JWT.RequiredClaims) == 0 {
		return fmt.Errorf("jwt.requiredClaims cannot be empty")
	}

	// Logger validation
	validLogLevels := []string{"trace", "debug", "info", "warn", "error", "fatal", "panic"}
	if !contains(validLogLevels, config.Logger.Level) {
		return fmt.Errorf("logger.level must be one of %v, got %s", validLogLevels, config.Logger.Level)
	}

	validLogFormats := []string{"json", "console"}
	if !contains(validLogFormats, config.Logger.Format) {
		return fmt.Errorf("logger.format must be one of %v, got %s", validLogFormats, config.Logger.Format)
	}

	// Environment validation
	validEnvironments := []string{"development", "staging", "production"}
	if !contains(validEnvironments, config.Environment) {
		return fmt.Errorf("environment must be one of %v, got %s", validEnvironments, config.Environment)
	}

	return nil
}

// contains checks if a slice contains a specific string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
