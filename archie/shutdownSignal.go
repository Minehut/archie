package archie

import (
	"context"
	"fmt"
	"github.com/rs/zerolog/log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func (a *Archiver) WaitForSignal(
	shutdownWait string,
	baseCancel context.CancelFunc,
	msgCancel context.CancelFunc,
	healthCheckSrv *http.Server,
	metricsSrv *http.Server,
) {
	shutdownWaitDuration, err := time.ParseDuration(shutdownWait)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to parse shutdown-wait duration argument")
	}

	if shutdownWaitDuration > 0 {
		log.Info().Msgf("Shutdown wait duration set to %s", shutdownWaitDuration)
	}

	log.Info().Msg("Startup complete")

	// wait for signals to end script
	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel, os.Interrupt, syscall.SIGTERM)

	for {
		// blocking
		sig := <-signalChannel
		switch sig {
		case os.Interrupt:
			fmt.Println()

			// double ctrl-c exits immediately
			go func() {
				<-signalChannel
				fmt.Println()
				os.Exit(99)
			}()

			a.shutdown(sig, shutdownWaitDuration, baseCancel, msgCancel, healthCheckSrv, metricsSrv)
			return
		case syscall.SIGTERM:
			a.shutdown(sig, shutdownWaitDuration, baseCancel, msgCancel, healthCheckSrv, metricsSrv)
			return
		}
	}
}

func (a *Archiver) shutdown(
	sig os.Signal,
	shutdownWaitDuration time.Duration,
	baseCancel context.CancelFunc,
	msgCancel context.CancelFunc,
	healthCheckSrv *http.Server,
	metricsSrv *http.Server,
) {
	log.Info().Msgf("Received %s signal, starting %s shutdown wait", sig, shutdownWaitDuration)

	// stop fetching new records
	baseCancel()
	// wait for running transfers to finish
	// shutdownWaitForced to allow extra time for metrics to be scraped
	a.shutdownWaiter(shutdownWaitDuration, true)
	// cancel active transfers
	msgCancel()
	// stop health check server
	httpServerShutdown(healthCheckSrv)
	// stop metrics server
	httpServerShutdown(metricsSrv)
	// wait for everything to finish
	a.WaitGroup.Wait()
}

func (a *Archiver) shutdownWaiter(shutdownWaitDuration time.Duration, shutdownWaitForced bool) {
	fetchThreadDone := false
	// build a new context just for a timeout
	shutdownWaitCtx, shutdownWaitCancel := context.WithTimeout(context.Background(), shutdownWaitDuration)
	defer shutdownWaitCancel()

	for {
		// check for shutdown wait timeout expiration
		if checkContextDone(shutdownWaitCtx) {
			if fetchThreadDone {
				log.Info().Msg("Shutdown wait timeout has expired")
			} else {
				log.Info().Msg("Shutdown wait timeout has expired, running transfer will be terminated")
			}
			return
		}
		// check if message processor's jetstream fetch thread finished gracefully
		select {
		case m := <-a.FetchDone:
			log.Info().Msgf("All transfers have completed, %s shutdown of message thread", m)
			fetchThreadDone = true
			if !shutdownWaitForced {
				return
			}
		default:
		}
		time.Sleep(250 * time.Millisecond)
	}
}

func httpServerShutdown(httpServer *http.Server) {
	if httpServer != nil {
		// give the http server half a second max to shut down gracefully
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 500*time.Millisecond)

		defer func() {
			log.Trace().Msg("Deferred http server shutdown context canceled")
			shutdownCancel()
		}()

		err := httpServer.Shutdown(shutdownCtx)
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to shutdown the http server server")
		}
	}
}
