package execution

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"time"

	"github.com/rule-engine/internal/config"
	"github.com/rule-engine/internal/infra"
	"github.com/rule-engine/internal/infra/logger"
)

// CacheRefreshService orchestrates the cache refresh loop
type CacheRefreshService struct {
	engine         *Engine
	db             *sql.DB
	config         *config.Config
	logger         *logger.Logger
	stopChan       chan struct{}
	wg             sync.WaitGroup
	circuitBreaker *CircuitBreaker
	lastChecksums  map[string]string // Track checksums per namespace
	mu             sync.RWMutex
}

// CircuitBreaker implements the circuit breaker pattern for resilience
type CircuitBreaker struct {
	failures    int
	maxFailures int
	timeout     time.Duration
	lastFailure time.Time
	state       CircuitBreakerState
	mu          sync.RWMutex
}

// CircuitBreakerState represents the current state of the circuit breaker
type CircuitBreakerState int

const (
	CircuitBreakerClosed CircuitBreakerState = iota
	CircuitBreakerOpen
	CircuitBreakerHalfOpen
)

// Errors
var (
	ErrCircuitBreakerOpen = fmt.Errorf("circuit breaker is open")
	ErrServiceStopped     = fmt.Errorf("cache refresh service is stopped")
)

// NewCacheRefreshService creates a new cache refresh service
func NewCacheRefreshService(
	engine *Engine,
	db *sql.DB,
	config *config.Config,
	logger *logger.Logger,
) *CacheRefreshService {
	return &CacheRefreshService{
		engine:   engine,
		db:       db,
		config:   config,
		logger:   logger,
		stopChan: make(chan struct{}),
		circuitBreaker: &CircuitBreaker{
			maxFailures: config.CacheRefresh.CircuitBreaker.MaxFailures,
			timeout:     config.CacheRefresh.CircuitBreaker.Timeout,
			state:       CircuitBreakerClosed,
		},
		lastChecksums: make(map[string]string),
	}
}

// Start initializes and starts the refresh loop
func (c *CacheRefreshService) Start(ctx context.Context) error {
	if !c.config.CacheRefresh.Enabled {
		c.logger.Info().Msg("Cache refresh service is disabled")
		return nil
	}

	c.logger.Info().
		Dur("polling_interval", c.config.CacheRefresh.PollingInterval).
		Int("max_failures", c.config.CacheRefresh.CircuitBreaker.MaxFailures).
		Dur("timeout", c.config.CacheRefresh.CircuitBreaker.Timeout).
		Msg("Starting cache refresh service")

	// Initialize with current checksum
	if err := c.initializeChecksum(); err != nil {
		return fmt.Errorf("failed to initialize checksum: %w", err)
	}

	c.wg.Add(1)
	go c.refreshLoop(ctx)

	return nil
}

// Stop gracefully shuts down the refresh loop
func (c *CacheRefreshService) Stop() error {
	if !c.config.CacheRefresh.Enabled {
		return nil
	}

	c.logger.Info().Msg("Stopping cache refresh service")
	close(c.stopChan)
	c.wg.Wait()
	c.logger.Info().Msg("Cache refresh service stopped")
	return nil
}

// refreshLoop is the main polling loop
func (c *CacheRefreshService) refreshLoop(ctx context.Context) {
	defer c.wg.Done()

	ticker := time.NewTicker(c.config.CacheRefresh.PollingInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			c.logger.Info().Msg("Context cancelled, stopping cache refresh loop")
			return
		case <-c.stopChan:
			c.logger.Info().Msg("Stop signal received, stopping cache refresh loop")
			return
		case <-ticker.C:
			if err := c.checkForChanges(ctx); err != nil {
				c.logger.Error().Err(err).Msg("Failed to check for changes")
			}
		}
	}
}

// checkForChanges compares current database checksum with cached checksum
func (c *CacheRefreshService) checkForChanges(ctx context.Context) error {
	return c.circuitBreaker.Execute(func() error {
		// Get all namespaces that need to be checked
		namespaces, err := c.getNamespaces(ctx)
		if err != nil {
			return fmt.Errorf("failed to get namespaces: %w", err)
		}

		// Check each namespace for changes
		for _, namespace := range namespaces {
			if err := c.checkNamespaceChanges(ctx, namespace); err != nil {
				return fmt.Errorf("failed to check namespace %s: %w", namespace, err)
			}
		}

		return nil
	})
}

// checkNamespaceChanges checks for changes in a specific namespace
func (c *CacheRefreshService) checkNamespaceChanges(ctx context.Context, namespace string) error {
	// First, refresh the namespace checksum
	if err := c.refreshNamespaceChecksum(ctx, namespace); err != nil {
		infra.CacheRefreshErrors.WithLabelValues(namespace).Inc()
		return fmt.Errorf("failed to refresh namespace checksum: %w", err)
	}

	// Get the current checksum from the database
	currentChecksum, err := c.getNamespaceChecksum(ctx, namespace)
	if err != nil {
		infra.CacheRefreshErrors.WithLabelValues(namespace).Inc()
		return fmt.Errorf("failed to get namespace checksum: %w", err)
	}

	c.mu.RLock()
	lastChecksum := c.lastChecksums[namespace]
	c.mu.RUnlock()

	if currentChecksum != lastChecksum {
		c.logger.Info().
			Str("namespace", namespace).
			Str("old_checksum", lastChecksum).
			Str("new_checksum", currentChecksum).
			Msg("Namespace changes detected, refreshing cache")

		if err := c.refreshCache(ctx, namespace); err != nil {
			infra.CacheRefreshErrors.WithLabelValues(namespace).Inc()
			return fmt.Errorf("failed to refresh cache: %w", err)
		}

		c.mu.Lock()
		c.lastChecksums[namespace] = currentChecksum
		c.mu.Unlock()

		// Update metrics
		now := time.Now()
		infra.CacheRefreshLastTime.WithLabelValues(namespace).Set(float64(now.Unix()))
		infra.CacheRefreshStaleness.WithLabelValues(namespace).Set(0) // Reset staleness after refresh

		c.logger.Info().Str("namespace", namespace).Msg("Cache refreshed successfully")
	} else {
		// Update staleness metric for unchanged cache
		lastRefresh := c.getLastRefreshTime(namespace)
		if !lastRefresh.IsZero() {
			staleness := time.Since(lastRefresh).Seconds()
			infra.CacheRefreshStaleness.WithLabelValues(namespace).Set(staleness)
		}
	}

	return nil
}

// refreshNamespaceChecksum calls the refresh_namespace_checksum function
func (c *CacheRefreshService) refreshNamespaceChecksum(ctx context.Context, namespace string) error {
	_, err := c.db.ExecContext(ctx, "SELECT refresh_namespace_checksum($1)", namespace)
	if err != nil {
		return fmt.Errorf("failed to refresh namespace checksum: %w", err)
	}
	return nil
}

// getNamespaceChecksum retrieves the current namespace checksum
func (c *CacheRefreshService) getNamespaceChecksum(ctx context.Context, namespace string) (string, error) {
	var checksum string
	err := c.db.QueryRowContext(ctx, "SELECT checksum FROM active_config_meta WHERE namespace = $1", namespace).Scan(&checksum)
	if err != nil {
		return "", fmt.Errorf("failed to query namespace checksum: %w", err)
	}
	return checksum, nil
}

// getCurrentChecksum retrieves the current database checksum
func (c *CacheRefreshService) getCurrentChecksum(ctx context.Context) (string, error) {
	// This function is no longer needed since we're using per-namespace checksums
	// Keeping it for backward compatibility but it should not be used
	return "", fmt.Errorf("getCurrentChecksum is deprecated, use getNamespaceChecksum instead")
}

// initializeChecksum sets the initial checksum value
func (c *CacheRefreshService) initializeChecksum() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Get all namespaces and initialize their checksums
	namespaces, err := c.getNamespaces(ctx)
	if err != nil {
		return fmt.Errorf("failed to get namespaces: %w", err)
	}

	// For now, we'll use a simple approach - just refresh all namespace checksums
	// In a production environment, you might want to be more selective
	for _, namespace := range namespaces {
		if err := c.refreshNamespaceChecksum(ctx, namespace); err != nil {
			c.logger.Warn().Str("namespace", namespace).Err(err).Msg("Failed to initialize namespace checksum")
		}
	}

	c.logger.Info().Int("namespace_count", len(namespaces)).Msg("Initialized cache refresh service")
	return nil
}

// getNamespaces retrieves all active namespaces
func (c *CacheRefreshService) getNamespaces(ctx context.Context) ([]string, error) {
	rows, err := c.db.QueryContext(ctx, "SELECT id FROM namespaces")
	if err != nil {
		return nil, fmt.Errorf("failed to query namespaces: %w", err)
	}
	defer rows.Close()

	var namespaces []string
	for rows.Next() {
		var namespace string
		if err := rows.Scan(&namespace); err != nil {
			return nil, fmt.Errorf("failed to scan namespace: %w", err)
		}
		namespaces = append(namespaces, namespace)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating namespaces: %w", err)
	}

	return namespaces, nil
}

// refreshCache refreshes the engine cache for a specific namespace
func (c *CacheRefreshService) refreshCache(ctx context.Context, namespace string) error {
	startTime := time.Now()

	if err := c.engine.RefreshNamespaceCache(ctx, namespace); err != nil {
		return fmt.Errorf("failed to refresh namespace cache: %w", err)
	}

	// Update metrics for successful refresh
	refreshDuration := time.Since(startTime).Seconds()
	c.logger.Debug().
		Float64("duration_seconds", refreshDuration).
		Msg("Cache refresh completed")

	return nil
}

// getLastRefreshTime returns the last refresh time for a namespace
func (c *CacheRefreshService) getLastRefreshTime(namespace string) time.Time {
	// This would need to be tracked in the service
	// For now, return zero time
	return time.Time{}
}

// Circuit Breaker Implementation

// Execute runs an operation through the circuit breaker
func (cb *CircuitBreaker) Execute(operation func() error) error {
	cb.mu.RLock()
	if cb.isOpen() {
		cb.mu.RUnlock()
		return ErrCircuitBreakerOpen
	}
	cb.mu.RUnlock()

	err := operation()
	cb.recordResult(err)
	return err
}

// isOpen checks if the circuit breaker is in open state
func (cb *CircuitBreaker) isOpen() bool {
	if cb.state == CircuitBreakerOpen {
		// Check if timeout has passed to transition to half-open
		if time.Since(cb.lastFailure) >= cb.timeout {
			cb.state = CircuitBreakerHalfOpen
			return false
		}
		return true
	}
	return false
}

// recordResult records the result of an operation
func (cb *CircuitBreaker) recordResult(err error) {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	if err != nil {
		cb.failures++
		cb.lastFailure = time.Now()

		if cb.failures >= cb.maxFailures {
			cb.state = CircuitBreakerOpen
		}
	} else {
		// Success - reset failures and close circuit breaker
		cb.failures = 0
		cb.state = CircuitBreakerClosed
	}
}

// GetState returns the current circuit breaker state
func (cb *CircuitBreaker) GetState() CircuitBreakerState {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.state
}

// GetFailures returns the current failure count
func (cb *CircuitBreaker) GetFailures() int {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.failures
}

// GetLastFailure returns the timestamp of the last failure
func (cb *CircuitBreaker) GetLastFailure() time.Time {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.lastFailure
}
