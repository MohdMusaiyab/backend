package telemetry

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// NotificationsProcessed is a Counter vector. It strictly counts up.
// The labels "channel" (email/sms) and "status" (success/failed) allow us to filter in Grafana.
var NotificationsProcessed = promauto.NewCounterVec(
	prometheus.CounterOpts{
		Name: "notification_processed_total",
		Help: "The total number of processed notifications",
	},
	[]string{"channel", "status"},
)

// NotificationLatency is a Histogram. It records how long specific actions take,
// allowing Prometheus to calculate the p50, p95, and p99 latency percentiles!
var NotificationLatency = promauto.NewHistogramVec(
	prometheus.HistogramOpts{
		Name:    "notification_processing_duration_seconds",
		Help:    "Histogram of processing time for notifications in seconds",
		Buckets: prometheus.DefBuckets, // Uses standard buckets (e.g. 0.005s to 10s)
	},
	[]string{"channel"},
)
