package foundation

import (
	"fmt"
	"io"
	"net/http"

	"github.com/rs/zerolog/log"
)

// InitLivenessAndReadiness initializes the /liveness and /readiness endpoint on port 5000
func InitLivenessAndReadiness() {
	InitLivenessAndReadinessWithPort(5000)
}

// InitLivenessAndReadinessWithPort initializes the /liveness and /readiness endpoint on specified port
func InitLivenessAndReadinessWithPort(port int) {
	// start liveness endpoint
	go func() {
		portString := fmt.Sprintf(":%v", port)
		log.Debug().
			Str("port", portString).
			Msg("Serving /liveness and /readiness endpoints...")

		serverMux := http.NewServeMux()
		serverMux.HandleFunc("/liveness", func(w http.ResponseWriter, _ *http.Request) {
			io.WriteString(w, "I'm alive!\n")
		})
		serverMux.HandleFunc("/readiness", func(w http.ResponseWriter, _ *http.Request) {
			io.WriteString(w, "I'm ready!\n")
		})

		if err := http.ListenAndServe(portString, serverMux); err != nil {
			log.Fatal().Err(err).Msg("Starting /liveness and /readiness listener failed")
		}
	}()
}
