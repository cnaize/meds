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

	packetsAcceptCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "meds",
			Subsystem: "core",
			Name:      "packets_accept_count",
			Help:      "Total number of accepted packets",
		},
		[]string{"reason", "filter"},
	)

	packetsDropCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "meds",
			Subsystem: "core",
			Name:      "packets_drop_count",
			Help:      "Total number of dropped packets",
		},
		[]string{"reason", "filter"},
	)
)

func init() {
	prometheus.MustRegister(packetsTotalCounter)
	prometheus.MustRegister(packetsAcceptCounter)
	prometheus.MustRegister(packetsDropCounter)
}
