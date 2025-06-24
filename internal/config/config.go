package config

import (
	"fmt"
	"time"
)

// Config represents the complete application configuration
type Config struct {
	Environment  string             `yaml:"environment" mapstructure:"environment"`
	Server       ServerConfig       `yaml:"server" mapstructure:"server"`
	Database     DatabaseConfig     `yaml:"database" mapstructure:"database"`
	Cache        CacheConfig        `yaml:"cache" mapstructure:"cache"`
	CacheRefresh CacheRefreshConfig `yaml:"cache_refresh" mapstructure:"cache_refresh"`
	JWT          JWTConfig          `yaml:"jwt" mapstructure:"jwt"`
	Metrics      MetricsConfig      `yaml:"metrics" mapstructure:"metrics"`
	Logger       LoggerConfig       `yaml:"logger" mapstructure:"logger"`
}

// ServerConfig holds HTTP server configuration
type ServerConfig struct {
	Port            int           `yaml:"port" mapstructure:"port"`
	Host            string        `yaml:"host" mapstructure:"host"`
	ReadTimeout     time.Duration `yaml:"readTimeout" mapstructure:"readTimeout"`
	WriteTimeout    time.Duration `yaml:"writeTimeout" mapstructure:"writeTimeout"`
	IdleTimeout     time.Duration `yaml:"idleTimeout" mapstructure:"idleTimeout"`
	ShutdownTimeout time.Duration `yaml:"shutdownTimeout" mapstructure:"shutdownTimeout"`
	EnableCORS      bool          `yaml:"enableCORS" mapstructure:"enableCORS"`
	TrustedProxies  []string      `yaml:"trustedProxies" mapstructure:"trustedProxies"`
}

// DatabaseConfig holds database connection configuration
type DatabaseConfig struct {
	Host            string        `yaml:"host" mapstructure:"host"`
	Port            int           `yaml:"port" mapstructure:"port"`
	Name            string        `yaml:"name" mapstructure:"name"`
	User            string        `yaml:"user" mapstructure:"user"`
	Password        string        `yaml:"password" mapstructure:"password"`
	SSLMode         string        `yaml:"sslMode" mapstructure:"sslMode"`
	MaxConnections  int           `yaml:"maxConnections" mapstructure:"maxConnections"`
	MaxIdleConns    int           `yaml:"maxIdleConns" mapstructure:"maxIdleConns"`
	ConnMaxLifetime time.Duration `yaml:"connMaxLifetime" mapstructure:"connMaxLifetime"`
	QueryTimeout    time.Duration `yaml:"queryTimeout" mapstructure:"queryTimeout"`
	// URL allows override with full connection string
	URL string `yaml:"url" mapstructure:"url"`
}

// CacheConfig holds in-memory cache configuration
type CacheConfig struct {
	RefreshIntervalSec int                  `yaml:"refreshIntervalSec" mapstructure:"refreshIntervalSec"`
	RefreshJitterSec   int                  `yaml:"refreshJitterSec" mapstructure:"refreshJitterSec"`
	RefreshTimeout     time.Duration        `yaml:"refreshTimeout" mapstructure:"refreshTimeout"`
	MaxSizeMB          int                  `yaml:"maxSizeMB" mapstructure:"maxSizeMB"`
	CircuitBreaker     CircuitBreakerConfig `yaml:"circuitBreaker" mapstructure:"circuitBreaker"`
}

// CircuitBreakerConfig holds circuit breaker configuration for DB refresh calls
type CircuitBreakerConfig struct {
	Enabled          bool          `yaml:"enabled" mapstructure:"enabled"`
	FailureThreshold int           `yaml:"failureThreshold" mapstructure:"failureThreshold"`
	RecoveryTimeout  time.Duration `yaml:"recoveryTimeout" mapstructure:"recoveryTimeout"`
	MaxRequests      int           `yaml:"maxRequests" mapstructure:"maxRequests"`
}

// JWTConfig holds JWT authentication configuration
type JWTConfig struct {
	Secret          string        `yaml:"secret" mapstructure:"secret"`
	TokenExpiration time.Duration `yaml:"tokenExpiration" mapstructure:"tokenExpiration"`
	RequiredClaims  []string      `yaml:"requiredClaims" mapstructure:"requiredClaims"`
	ValidAudiences  []string      `yaml:"validAudiences" mapstructure:"validAudiences"`
	ValidIssuers    []string      `yaml:"validIssuers" mapstructure:"validIssuers"`
	SkipExpiryCheck bool          `yaml:"skipExpiryCheck" mapstructure:"skipExpiryCheck"`
}

// MetricsConfig holds Prometheus metrics configuration
type MetricsConfig struct {
	Enabled   bool   `yaml:"enabled" mapstructure:"enabled"`
	Path      string `yaml:"path" mapstructure:"path"`
	Port      int    `yaml:"port" mapstructure:"port"`
	Namespace string `yaml:"namespace" mapstructure:"namespace"`
	Subsystem string `yaml:"subsystem" mapstructure:"subsystem"`
}

// LoggerConfig holds logging configuration
type LoggerConfig struct {
	Level        string `yaml:"level" mapstructure:"level"`
	Format       string `yaml:"format" mapstructure:"format"` // json, console
	EnableCaller bool   `yaml:"enableCaller" mapstructure:"enableCaller"`
	TimeFormat   string `yaml:"timeFormat" mapstructure:"timeFormat"`
	LogPath      string `yaml:"logPath" mapstructure:"logPath"` // file path for logs, empty for stdout
}

// RateLimitConfig holds rate limiting configuration
type RateLimitConfig struct {
	Enabled         bool `yaml:"enabled" mapstructure:"enabled"`
	GlobalRPM       int  `yaml:"globalRPM" mapstructure:"globalRPM"`
	PerUserRPM      int  `yaml:"perUserRPM" mapstructure:"perUserRPM"`
	ExecutionRPM    int  `yaml:"executionRPM" mapstructure:"executionRPM"`
	BurstMultiplier int  `yaml:"burstMultiplier" mapstructure:"burstMultiplier"`
}

// CacheRefreshConfig holds cache refresh loop configuration
type CacheRefreshConfig struct {
	Enabled         bool                 `yaml:"enabled" mapstructure:"enabled"`
	PollingInterval time.Duration        `yaml:"polling_interval" mapstructure:"polling_interval"`
	CircuitBreaker  CacheRefreshCBConfig `yaml:"circuit_breaker" mapstructure:"circuit_breaker"`
}

// CacheRefreshCBConfig holds circuit breaker configuration for cache refresh
type CacheRefreshCBConfig struct {
	MaxFailures int           `yaml:"max_failures" mapstructure:"max_failures"`
	Timeout     time.Duration `yaml:"timeout" mapstructure:"timeout"`
}

// GetDatabaseURL returns the complete database connection URL
func (d *DatabaseConfig) GetDatabaseURL() string {
	// Use direct URL if provided
	if d.URL != "" {
		return d.URL
	}

	// Build URL from components
	return fmt.Sprintf("postgresql://%s:%s@%s:%d/%s?sslmode=%s",
		d.User, d.Password, d.Host, d.Port, d.Name, d.SSLMode)
}

// GetServerAddress returns the complete server address
func (s *ServerConfig) GetServerAddress() string {
	return fmt.Sprintf("%s:%d", s.Host, s.Port)
}

// IsProduction returns true if running in production environment
func (c *Config) IsProduction() bool {
	return c.Environment == "production"
}

// IsDevelopment returns true if running in development environment
func (c *Config) IsDevelopment() bool {
	return c.Environment == "development"
}

// IsStaging returns true if running in staging environment
func (c *Config) IsStaging() bool {
	return c.Environment == "staging"
}
