package foundation

import (
	"fmt"
	"io"
	"net/http"

	"github.com/rs/zerolog/log"
)

// InitLiveness initializes the /liveness endpoint on port 5000
func InitLiveness() {
	InitLivenessWithPort(5000)
}

// InitLivenessWithPort initializes the /liveness endpoint on specified port
func InitLivenessWithPort(port int) {
	// start liveness endpoint
	go func() {
		portString := fmt.Sprintf(":%v", port)
		log.Debug().
			Str("port", portString).
			Msg("Serving /liveness endpoint...")

		serverMux := http.NewServeMux()
		serverMux.HandleFunc("/liveness", func(w http.ResponseWriter, _ *http.Request) {
			io.WriteString(w, "I'm alive!\n")
		})

		if err := http.ListenAndServe(portString, serverMux); err != nil {
			log.Fatal().Err(err).Msg("Starting /liveness listener failed")
		}
	}()
}
