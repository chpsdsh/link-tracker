package metrics

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

var SentNotificationsTotal = prometheus.NewCounter(
	prometheus.CounterOpts{
		Name: "sent_notification_total",
		Help: "Total count of sent notifications",
		ConstLabels: prometheus.Labels{
			"app": "bot",
		},
	},
)

var CommandDurationMsTotal = prometheus.NewHistogramVec(
	prometheus.HistogramOpts{
		Name: "command_duration_ms_total",
		Help: "Duration of bot processing operations in milliseconds.",
		ConstLabels: prometheus.Labels{
			"app": "bot",
		},
		Buckets: []float64{10, 25, 50, 100, 250, 500, 1000, 2500, 5000},
	},
	[]string{"scope", "scope_type"},
)

var CommandRequestTotal = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "command_requests_total",
		Help: "Counter of bot's handled commands.",
		ConstLabels: prometheus.Labels{
			"app": "bot",
		},
	},
	[]string{"command"},
)

var ErrorsCounterTotal = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "errors_counter_total",
		Help: "Counter of bot's errors",
		ConstLabels: prometheus.Labels{
			"app": "bot",
		},
	},
	[]string{"scope", "scope_type"},
)

func RegisterBotMetrics() {
	prometheus.MustRegister(SentNotificationsTotal)
	prometheus.MustRegister(CommandDurationMsTotal)
	prometheus.MustRegister(CommandRequestTotal)
	prometheus.MustRegister(ErrorsCounterTotal)
}

func ObserveCommandDuration(start time.Time, scope string, scopeType string) {
	CommandDurationMsTotal.
		WithLabelValues(scope, scopeType).
		Observe(float64(time.Since(start).Milliseconds()))
}
