package metrics

import (
	"github.com/maypok86/otter/v2/stats"
	"github.com/prometheus/client_golang/prometheus"
)

type Metrics struct {
	PacketsAcceptedTotal  *prometheus.CounterVec
	PacketsDroppedTotal   *prometheus.CounterVec
	PacketsProcessedTotal prometheus.Counter
	ErrorsTotal           *prometheus.CounterVec
	RateLimiterCacheStats *stats.Counter
}

var metrics *Metrics

func init() {
	metrics = &Metrics{
		PacketsAcceptedTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "meds",
				Subsystem: "core",
				Name:      "packets_accepted_total",
				Help:      "Total number of accepted packets",
			},
			[]string{"reason", "filter"},
		),
		PacketsDroppedTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "meds",
				Subsystem: "core",
				Name:      "packets_dropped_total",
				Help:      "Total number of dropped packets",
			},
			[]string{"reason", "filter"},
		),
		PacketsProcessedTotal: prometheus.NewCounter(
			prometheus.CounterOpts{
				Namespace: "meds",
				Subsystem: "core",
				Name:      "packets_processed_total",
				Help:      "Total number of processed packets",
			},
		),
		ErrorsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "meds",
				Subsystem: "core",
				Name:      "errors_total",
				Help:      "Total number of errors",
			},
			[]string{"error"},
		),
		RateLimiterCacheStats: stats.NewCounter(),
	}
}

func Get() *Metrics {
	return metrics
}

func (m *Metrics) Register(reg *prometheus.Registry) {
	reg.MustRegister(m.PacketsAcceptedTotal)
	reg.MustRegister(m.PacketsDroppedTotal)
	reg.MustRegister(m.PacketsProcessedTotal)
	reg.MustRegister(m.ErrorsTotal)
}
