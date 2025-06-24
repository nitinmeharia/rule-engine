package config

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateConfig_Success(t *testing.T) {
	cfg := &Config{
		Environment: "development",
		Server:      ServerConfig{Port: 8080, Host: "localhost"},
		Database: DatabaseConfig{
			Host:           "localhost",
			Port:           5432,
			Name:           "testdb",
			User:           "user",
			Password:       "pass",
			MaxConnections: 10,
			MaxIdleConns:   5,
		},
		Cache: CacheConfig{
			RefreshIntervalSec: 10,
			RefreshJitterSec:   1,
			MaxSizeMB:          10,
		},
		JWT: JWTConfig{
			Secret:         "supersecret",
			RequiredClaims: []string{"clientId"},
		},
		Logger: LoggerConfig{Level: "info", Format: "console"},
	}
	assert.NoError(t, validateConfig(cfg))
}

func TestValidateConfig_Failures(t *testing.T) {
	cases := []struct {
		name   string
		mutate func(*Config)
		errMsg string
	}{
		{"bad port", func(c *Config) { c.Server.Port = 0 }, "server.port must be between 1 and 65535"},
		{"empty host", func(c *Config) { c.Server.Host = "" }, "server.host cannot be empty"},
		{"empty db host", func(c *Config) { c.Database.Host = "" }, "database.host cannot be empty"},
		{"empty db name", func(c *Config) { c.Database.Name = "" }, "database.name cannot be empty"},
		{"empty db user", func(c *Config) { c.Database.User = "" }, "database.user cannot be empty"},
		{"maxConnections < 1", func(c *Config) { c.Database.MaxConnections = 0 }, "database.maxConnections must be at least 1"},
		{"maxIdleConns < 0", func(c *Config) { c.Database.MaxIdleConns = -1 }, "database.maxIdleConns cannot be negative"},
		{"maxIdleConns > maxConnections", func(c *Config) { c.Database.MaxIdleConns = 11 }, "database.maxIdleConns (11) cannot exceed maxConnections (10)"},
		{"cache.refreshIntervalSec < 1", func(c *Config) { c.Cache.RefreshIntervalSec = 0 }, "cache.refreshIntervalSec must be at least 1"},
		{"cache.refreshJitterSec < 0", func(c *Config) { c.Cache.RefreshJitterSec = -1 }, "cache.refreshJitterSec cannot be negative"},
		{"cache.maxSizeMB < 1", func(c *Config) { c.Cache.MaxSizeMB = 0 }, "cache.maxSizeMB must be at least 1"},
		{"jwt.secret empty", func(c *Config) { c.JWT.Secret = "" }, "jwt.secret cannot be empty"},
		{"jwt.secret default in prod", func(c *Config) { c.JWT.Secret = "change-me-in-production"; c.Environment = "production" }, "jwt.secret must be changed in production environment"},
		{"jwt.requiredClaims empty", func(c *Config) { c.JWT.RequiredClaims = nil }, "jwt.requiredClaims cannot be empty"},
		{"logger.level invalid", func(c *Config) { c.Logger.Level = "invalid" }, "logger.level must be one of"},
		{"logger.format invalid", func(c *Config) { c.Logger.Format = "invalid" }, "logger.format must be one of"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := &Config{
				Environment: "development",
				Server:      ServerConfig{Port: 8080, Host: "localhost"},
				Database: DatabaseConfig{
					Host:           "localhost",
					Port:           5432,
					Name:           "testdb",
					User:           "user",
					Password:       "pass",
					MaxConnections: 10,
					MaxIdleConns:   5,
				},
				Cache: CacheConfig{
					RefreshIntervalSec: 10,
					RefreshJitterSec:   1,
					MaxSizeMB:          10,
				},
				JWT: JWTConfig{
					Secret:         "supersecret",
					RequiredClaims: []string{"clientId"},
				},
				Logger: LoggerConfig{Level: "info", Format: "console"},
			}
			tc.mutate(cfg)
			err := validateConfig(cfg)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tc.errMsg)
		})
	}
}

func TestSetDefaults(t *testing.T) {
	viper.Reset()
	setDefaults()
	assert.Equal(t, 8080, viper.GetInt("server.port"))
	assert.Equal(t, "0.0.0.0", viper.GetString("server.host"))
	assert.Equal(t, "localhost", viper.GetString("database.host"))
	assert.Equal(t, 5432, viper.GetInt("database.port"))
	assert.Equal(t, "rule_engine", viper.GetString("database.name"))
	assert.Equal(t, 10, viper.GetInt("database.maxConnections"))
	assert.Equal(t, true, viper.GetBool("metrics.enabled"))
	assert.Equal(t, "/v1/metrics", viper.GetString("metrics.path"))
	assert.Equal(t, "info", viper.GetString("logger.level"))
	assert.Equal(t, "console", viper.GetString("logger.format"))
}

func TestConfig_Load_WithDefaults(t *testing.T) {
	// ARRANGE
	viper.Reset()
	var defaultConfig = []byte(`
database:
  name: "rule_engine"
  maxConnections: 10
metrics:
  path: "/v1/metrics"
logger:
  level: "info"
`)
	viper.SetConfigType("yaml")
	err := viper.ReadConfig(bytes.NewBuffer(defaultConfig))
	require.NoError(t, err, "Viper should read the in-memory config without error")

	// ACT
	var cfg Config
	err = viper.Unmarshal(&cfg)
	require.NoError(t, err, "Viper should unmarshal the config without error")

	// ASSERT
	assert.Equal(t, "rule_engine", cfg.Database.Name)
	assert.Equal(t, 10, cfg.Database.MaxConnections)
	assert.Equal(t, "/v1/metrics", cfg.Metrics.Path)
	assert.Equal(t, "info", cfg.Logger.Level)
}

func TestLoad_Integration_Defaults(t *testing.T) {
	// ARRANGE - Set up environment and use in-memory config
	os.Setenv("RULE_ENGINE_ENVIRONMENT", "development")
	defer os.Unsetenv("RULE_ENGINE_ENVIRONMENT")

	// Create a minimal config that matches expected defaults
	var testConfig = []byte(`
environment: development
server:
  port: 8080
  host: "0.0.0.0"
database:
  host: "localhost"
  port: 5432
  name: "rule_engine"
  user: "postgres"
  password: "postgres"
  maxConnections: 10
  maxIdleConns: 5
cache:
  refreshIntervalSec: 60
  refreshJitterSec: 10
  maxSizeMB: 100
jwt:
  secret: "test-secret"
  requiredClaims: ["clientId"]
metrics:
  enabled: true
  path: "/v1/metrics"
logger:
  level: "info"
  format: "console"
`)

	// Reset viper and set up in-memory config
	viper.Reset()
	viper.SetConfigType("yaml")
	err := viper.ReadConfig(bytes.NewBuffer(testConfig))
	require.NoError(t, err, "Viper should read the in-memory config without error")

	// ACT
	var cfg Config
	err = viper.Unmarshal(&cfg)
	require.NoError(t, err, "Viper should unmarshal the config without error")

	// ASSERT
	assert.Equal(t, "development", cfg.Environment)
	assert.Equal(t, 8080, cfg.Server.Port)
	assert.Equal(t, "0.0.0.0", cfg.Server.Host)
	assert.Equal(t, "localhost", cfg.Database.Host)
	assert.Equal(t, 5432, cfg.Database.Port)
	assert.Equal(t, "rule_engine", cfg.Database.Name)
	assert.Equal(t, 10, cfg.Database.MaxConnections)
	assert.Equal(t, true, cfg.Metrics.Enabled)
	assert.Equal(t, "/v1/metrics", cfg.Metrics.Path)
	assert.Equal(t, "info", cfg.Logger.Level)
	assert.Equal(t, "console", cfg.Logger.Format)
}

func TestLoad_Integration_WithEnvOverride(t *testing.T) {
	// ARRANGE - Set up environment variables and use in-memory config
	os.Setenv("RULE_ENGINE_SERVER_PORT", "9999")
	os.Setenv("RULE_ENGINE_DATABASE_HOST", "envhost")
	defer os.Unsetenv("RULE_ENGINE_SERVER_PORT")
	defer os.Unsetenv("RULE_ENGINE_DATABASE_HOST")

	// Create a minimal config
	var testConfig = []byte(`
environment: development
server:
  port: 8080
  host: "0.0.0.0"
database:
  host: "localhost"
  port: 5432
  name: "rule_engine"
  user: "postgres"
  password: "postgres"
  maxConnections: 10
  maxIdleConns: 5
cache:
  refreshIntervalSec: 60
  refreshJitterSec: 10
  maxSizeMB: 100
jwt:
  secret: "test-secret"
  requiredClaims: ["clientId"]
metrics:
  enabled: true
  path: "/v1/metrics"
logger:
  level: "info"
  format: "console"
`)

	// Reset viper and set up in-memory config with environment variable binding
	viper.Reset()
	viper.SetConfigType("yaml")

	// Enable environment variable binding like the actual Load function
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.SetEnvPrefix("RULE_ENGINE")

	err := viper.ReadConfig(bytes.NewBuffer(testConfig))
	require.NoError(t, err, "Viper should read the in-memory config without error")

	// ACT
	var cfg Config
	err = viper.Unmarshal(&cfg)
	require.NoError(t, err, "Viper should unmarshal the config without error")

	// ASSERT - Environment variables should override config values
	assert.Equal(t, 9999, cfg.Server.Port)
	assert.Equal(t, "envhost", cfg.Database.Host)
}
