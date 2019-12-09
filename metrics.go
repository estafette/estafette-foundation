package foundation

import (
	"fmt"
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog/log"
)

// InitMetrics initializes the prometheus endpoint /metrics on port 9101
func InitMetrics() {
	InitMetricsWithPort(9101)
}

// InitMetricsWithPort initializes the prometheus endpoint /metrics on specified port
func InitMetricsWithPort(port int) {
	// start prometheus
	go func() {
		portString := fmt.Sprintf(":%v", port)
		log.Debug().
			Str("port", portString).
			Msg("Serving Prometheus metrics...")

		http.Handle("/metrics", promhttp.Handler())

		if err := http.ListenAndServe(portString, nil); err != nil {
			log.Fatal().Err(err).Msg("Starting Prometheus listener failed")
		}
	}()
}
