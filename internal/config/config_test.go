package config

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDatabaseConfig_GetDatabaseURL(t *testing.T) {
	cfg := DatabaseConfig{
		Host:     "localhost",
		Port:     5432,
		Name:     "testdb",
		User:     "testuser",
		Password: "testpass",
		SSLMode:  "disable",
	}
	url := cfg.GetDatabaseURL()
	expected := "postgresql://testuser:testpass@localhost:5432/testdb?sslmode=disable"
	assert.Equal(t, expected, url)
}

func TestDatabaseConfig_GetDatabaseURL_WithDirectURL(t *testing.T) {
	directURL := "postgresql://customuser:custompass@customhost:5433/customdb?sslmode=require"
	cfg := DatabaseConfig{
		URL: directURL,
	}
	url := cfg.GetDatabaseURL()
	assert.Equal(t, directURL, url)
}

func TestServerConfig_GetServerAddress(t *testing.T) {
	s := ServerConfig{Host: "127.0.0.1", Port: 8080}
	assert.Equal(t, "127.0.0.1:8080", s.GetServerAddress())
}

func TestConfig_EnvironmentChecks(t *testing.T) {
	c := Config{Environment: "production"}
	assert.True(t, c.IsProduction())
	assert.False(t, c.IsDevelopment())
	assert.False(t, c.IsStaging())

	c.Environment = "development"
	assert.False(t, c.IsProduction())
	assert.True(t, c.IsDevelopment())
	assert.False(t, c.IsStaging())

	c.Environment = "staging"
	assert.False(t, c.IsProduction())
	assert.False(t, c.IsDevelopment())
	assert.True(t, c.IsStaging())

	c.Environment = "other"
	assert.False(t, c.IsProduction())
	assert.False(t, c.IsDevelopment())
	assert.False(t, c.IsStaging())
}

func TestCacheConfig_Defaults(t *testing.T) {
	cfg := CacheConfig{}
	assert.Equal(t, 0, cfg.RefreshIntervalSec)
	assert.Equal(t, 0, cfg.RefreshJitterSec)
	assert.Equal(t, time.Duration(0), cfg.RefreshTimeout)
	assert.Equal(t, 0, cfg.MaxSizeMB)
}

func TestJWTConfig_Defaults(t *testing.T) {
	cfg := JWTConfig{}
	assert.Equal(t, "", cfg.Secret)
	assert.Equal(t, time.Duration(0), cfg.TokenExpiration)
	assert.Nil(t, cfg.RequiredClaims)
	assert.Nil(t, cfg.ValidAudiences)
	assert.Nil(t, cfg.ValidIssuers)
	assert.False(t, cfg.SkipExpiryCheck)
}

func TestMetricsConfig_Defaults(t *testing.T) {
	cfg := MetricsConfig{}
	assert.False(t, cfg.Enabled)
	assert.Equal(t, "", cfg.Path)
	assert.Equal(t, 0, cfg.Port)
	assert.Equal(t, "", cfg.Namespace)
	assert.Equal(t, "", cfg.Subsystem)
}

func TestLoggerConfig_Defaults(t *testing.T) {
	cfg := LoggerConfig{}
	assert.Equal(t, "", cfg.Level)
	assert.Equal(t, "", cfg.Format)
	assert.False(t, cfg.EnableCaller)
	assert.Equal(t, "", cfg.TimeFormat)
	assert.Equal(t, "", cfg.LogPath)
}

func TestRateLimitConfig_Defaults(t *testing.T) {
	cfg := RateLimitConfig{}
	assert.False(t, cfg.Enabled)
	assert.Equal(t, 0, cfg.GlobalRPM)
	assert.Equal(t, 0, cfg.PerUserRPM)
	assert.Equal(t, 0, cfg.ExecutionRPM)
	assert.Equal(t, 0, cfg.BurstMultiplier)
}

func TestCacheRefreshConfig_Defaults(t *testing.T) {
	cfg := CacheRefreshConfig{}
	assert.False(t, cfg.Enabled)
	assert.Equal(t, time.Duration(0), cfg.PollingInterval)
	assert.Equal(t, 0, cfg.CircuitBreaker.MaxFailures)
	assert.Equal(t, time.Duration(0), cfg.CircuitBreaker.Timeout)
}

func TestCircuitBreakerConfig_Defaults(t *testing.T) {
	cfg := CircuitBreakerConfig{}
	assert.False(t, cfg.Enabled)
	assert.Equal(t, 0, cfg.FailureThreshold)
	assert.Equal(t, time.Duration(0), cfg.RecoveryTimeout)
	assert.Equal(t, 0, cfg.MaxRequests)
}
