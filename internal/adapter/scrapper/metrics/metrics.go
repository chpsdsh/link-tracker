package metrics

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

var LinksOnTrackTotal = prometheus.NewGaugeVec(
	prometheus.GaugeOpts{
		Name: "links_on_track_total",
		Help: "Total count of links on tracking by source.",
		ConstLabels: prometheus.Labels{
			"app": "scrapper",
		},
	},
	[]string{"tracked_source"},
)

var RequestDurationMsTotal = prometheus.NewHistogramVec(
	prometheus.HistogramOpts{
		Name: "request_duration_ms_total",
		Help: "Duration of bot command processing operations in milliseconds.",
		ConstLabels: prometheus.Labels{
			"app": "scrapper",
		},
		Buckets: []float64{10, 25, 50, 100, 250, 500, 1000, 2500, 5000},
	},
	[]string{"scope", "scope_type"},
)

var APIRequestsTotal = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "api_requests_total",
		Help: "Counter of API requests.",
		ConstLabels: prometheus.Labels{
			"app": "scrapper",
		},
	},
	[]string{"source"},
)

var APIErrorsTotal = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "api_errors_total",
		Help: "Counter of API errors.",
		ConstLabels: prometheus.Labels{
			"app": "scrapper",
		},
	},
	[]string{"source"},
)

func RegisterScrapperMetrics() {
	prometheus.MustRegister(LinksOnTrackTotal)
	prometheus.MustRegister(RequestDurationMsTotal)
	prometheus.MustRegister(APIRequestsTotal)
	prometheus.MustRegister(APIErrorsTotal)
}

func ObserveRequestDuration(start time.Time, scope string, scopeType string) {
	RequestDurationMsTotal.
		WithLabelValues(scope, scopeType).
		Observe(float64(time.Since(start).Milliseconds()))
}
