package event

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	packetsTotalCounter = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: "meds",
			Subsystem: "core",
			Name:      "packets_total_count",
			Help:      "Total number of processed packets",
		},
	)

	packetsAccetCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "meds",
			Subsystem: "core",
			Name:      "packets_accept_count",
			Help:      "Total number of accepted packets",
		},
		[]string{"action", "reason", "filter"},
	)

	packetsDropCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "meds",
			Subsystem: "core",
			Name:      "packets_drop_count",
			Help:      "Total number of dropped packets",
		},
		[]string{"action", "reason", "filter"},
	)
)

func init() {
	prometheus.MustRegister(packetsTotalCounter)
	prometheus.MustRegister(packetsAccetCounter)
	prometheus.MustRegister(packetsDropCounter)
}
