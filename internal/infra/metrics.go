package infra

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	CacheRefreshLastTime = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "cache_refresh_last_time",
			Help: "Last cache refresh time (unix timestamp) per namespace.",
		},
		[]string{"namespace"},
	)

	CacheRefreshStaleness = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "cache_refresh_staleness_seconds",
			Help: "Cache staleness in seconds per namespace.",
		},
		[]string{"namespace"},
	)

	CacheRefreshErrors = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "cache_refresh_errors_total",
			Help: "Total cache refresh errors per namespace.",
		},
		[]string{"namespace"},
	)

	ExecutionDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "execution_duration_seconds",
			Help:    "Execution duration in seconds per namespace.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"namespace"},
	)

	ExecutionErrors = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "execution_errors_total",
			Help: "Total execution errors per namespace.",
		},
		[]string{"namespace"},
	)
)

func RegisterMetrics() {
	prometheus.MustRegister(CacheRefreshLastTime)
	prometheus.MustRegister(CacheRefreshStaleness)
	prometheus.MustRegister(CacheRefreshErrors)
	prometheus.MustRegister(ExecutionDuration)
	prometheus.MustRegister(ExecutionErrors)
}
