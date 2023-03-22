package collector

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	crautiOpsProcessed = promauto.NewCounter(prometheus.CounterOpts{
		Name: "crauti_processed_ops_total",
		Help: "The total number of processed events",
	})
)
