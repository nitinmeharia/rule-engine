package infra

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
)

func TestRegisterMetrics(t *testing.T) {
	// Create a new registry for testing
	registry := prometheus.NewRegistry()

	// Register metrics to the test registry instead of default
	registry.MustRegister(CacheRefreshLastTime)
	registry.MustRegister(CacheRefreshStaleness)
	registry.MustRegister(CacheRefreshErrors)
	registry.MustRegister(ExecutionDuration)
	registry.MustRegister(ExecutionErrors)
}

func TestCacheRefreshLastTime_Set(t *testing.T) {
	// Create a new registry for this test
	registry := prometheus.NewRegistry()
	registry.MustRegister(CacheRefreshLastTime)

	CacheRefreshLastTime.WithLabelValues("ns1").Set(12345)
	val := testutil.ToFloat64(CacheRefreshLastTime.WithLabelValues("ns1"))
	if val != 12345 {
		t.Errorf("expected 12345, got %v", val)
	}
}

func TestCacheRefreshStaleness_Set(t *testing.T) {
	// Create a new registry for this test
	registry := prometheus.NewRegistry()
	registry.MustRegister(CacheRefreshStaleness)

	CacheRefreshStaleness.WithLabelValues("ns1").Set(42)
	val := testutil.ToFloat64(CacheRefreshStaleness.WithLabelValues("ns1"))
	if val != 42 {
		t.Errorf("expected 42, got %v", val)
	}
}

func TestCacheRefreshErrors_Inc(t *testing.T) {
	// Create a new registry for this test
	registry := prometheus.NewRegistry()
	registry.MustRegister(CacheRefreshErrors)

	CacheRefreshErrors.WithLabelValues("ns1").Inc()
	val := testutil.ToFloat64(CacheRefreshErrors.WithLabelValues("ns1"))
	if val != 1 {
		t.Errorf("expected 1, got %v", val)
	}
}

func TestExecutionDuration_Observe(t *testing.T) {
	// Create a new registry for this test
	registry := prometheus.NewRegistry()
	registry.MustRegister(ExecutionDuration)

	ExecutionDuration.WithLabelValues("ns1").Observe(2.5)
	count := testutil.CollectAndCount(ExecutionDuration)
	if count == 0 {
		t.Errorf("expected ExecutionDuration to have observations")
	}
}

func TestExecutionErrors_Inc(t *testing.T) {
	// Create a new registry for this test
	registry := prometheus.NewRegistry()
	registry.MustRegister(ExecutionErrors)

	ExecutionErrors.WithLabelValues("ns1").Inc()
	val := testutil.ToFloat64(ExecutionErrors.WithLabelValues("ns1"))
	if val != 1 {
		t.Errorf("expected 1, got %v", val)
	}
}
