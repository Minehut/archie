package archie

import (
	"fmt"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog/log"
	"net/http"
)

func (a *Archiver) StartMetricsServer(port int) *http.Server {

	srv := &http.Server{Addr: fmt.Sprintf(":%d", port)}

	http.Handle("/metrics", promhttp.Handler())

	go func() {
		defer func() {
			log.Trace().Msg("Deferred metrics wait group context canceled")
			a.WaitGroup.Done()
		}()
		a.WaitGroup.Add(1)

		// blocking
		err := srv.ListenAndServe()
		// if not a graceful close
		if err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("Failed to start the metrics server")
		}
	}()

	log.Info().Msgf("Started HTTP server for metrics listening on :%d", port)

	return srv
}
