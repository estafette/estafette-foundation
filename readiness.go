package foundation

import (
	"fmt"
	"io"
	"net/http"

	"github.com/rs/zerolog/log"
)

// InitReadiness initializes the /readiness endpoint on port 5000
func InitReadiness() {
	InitReadinessWithPort(5000)
}

// InitReadinessWithPort initializes the /readiness endpoint on specified port
func InitReadinessWithPort(port int) {
	// start liveness endpoint
	go func() {
		portString := fmt.Sprintf(":%v", port)
		log.Debug().
			Str("port", portString).
			Msg("Serving /readiness endpoint...")

		serverMux := http.NewServeMux()
		serverMux.HandleFunc("/readiness", func(w http.ResponseWriter, _ *http.Request) {
			io.WriteString(w, "I'm ready!\n")
		})

		if err := http.ListenAndServe(portString, serverMux); err != nil {
			log.Fatal().Err(err).Msg("Starting /readiness listener failed")
		}
	}()
}
