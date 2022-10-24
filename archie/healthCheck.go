package archie

import (
	"encoding/json"
	"fmt"
	"github.com/InVisionApp/go-health/v2"
	"github.com/InVisionApp/go-health/v2/handlers"
	"github.com/minio/minio-go/v7"
	"github.com/nats-io/nats.go"
	"github.com/rs/zerolog/log"

	"net/http"
	"time"
)

// HealthCheckStatusListener create our own status listener, so we can use zerolog
type HealthCheckStatusListener struct{}

type readinessCheck struct {
	DestClient    *minio.Client
	SrcClient     *minio.Client
	jetStreamConn *nats.Conn
}

type livenessCheck struct{}

func (a *Archiver) StartHealthCheckServer(healthCheckPort int, jetStreamConn *nats.Conn) *http.Server {
	if a.HealthCheckDisabled {
		return nil
	}

	// live
	cLivenessCheck := &livenessCheck{}
	livenessHandler := startHealthCheck("live", cLivenessCheck)

	// ready
	cReadinessCheck := &readinessCheck{
		SrcClient:     a.SrcClient,
		DestClient:    a.DestClient,
		jetStreamConn: jetStreamConn,
	}
	readinessHandler := startHealthCheck("ready", cReadinessCheck)

	srv := &http.Server{Addr: fmt.Sprintf(":%d", healthCheckPort)}

	http.Handle("/ready", handlers.NewJSONHandlerFunc(readinessHandler, nil))
	http.Handle("/live", handlers.NewJSONHandlerFunc(livenessHandler, nil))

	go func() {
		defer func() {
			log.Trace().Msg("Deferred health check wait group context canceled")
			a.WaitGroup.Done()
		}()
		a.WaitGroup.Add(1)

		// blocking
		err := srv.ListenAndServe()
		// if not a graceful close
		if err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("Failed to start the health check server")
		}
	}()

	log.Info().Msgf("Started HTTP server for health checks listening on :%d", healthCheckPort)

	return srv
}

func startHealthCheck(name string, custom health.ICheckable) *health.Health {
	healthCheck := health.New()
	// we use HealthCheckStatusListener{}
	// so that we can use zerolog there
	healthCheck.DisableLogging()

	err := healthCheck.AddChecks([]*health.Config{
		{
			Name:     name,
			Checker:  custom,
			Interval: time.Duration(2) * time.Second,
			Fatal:    true,
		},
	})
	if err != nil {
		log.Fatal().Msgf("Unable to add %s health check: %v", name, err)
	}

	// set custom status listener
	statusListener := &HealthCheckStatusListener{}
	healthCheck.StatusListener = statusListener

	err = healthCheck.Start()
	if err != nil {
		log.Fatal().Msgf("Unable to start %s health check: %v", name, err)
	}

	return healthCheck
}

func (l *livenessCheck) Status() (interface{}, error) {
	// TODO: add some internal checks here

	// You can return additional information pertaining to the check as long
	// as it can be JSON marshalled
	//return map[string]int{"foo": 123, "bar": 456}, nil
	return nil, nil
}
func (c *readinessCheck) Status() (interface{}, error) {

	if c.SrcClient.IsOffline() {
		return nil, fmt.Errorf("source client health check failed")
	}

	if c.DestClient.IsOffline() {
		return nil, fmt.Errorf("destination client health check failed")
	}

	if !c.jetStreamConn.IsConnected() {
		return nil, fmt.Errorf("jetstream client is not connected")
	}

	return nil, nil
}

// HealthCheckFailed is triggered when a health check fails the first time
func (sl *HealthCheckStatusListener) HealthCheckFailed(status *health.State) {
	statusJSON, err := json.Marshal(status)
	if err != nil {
		log.Error().Err(err).Msgf("Failed to parse health check status")
		log.Info().Msgf("Healthcheck status changed: %+v", status)
	} else {
		log.Info().RawJSON("status", statusJSON).Msg("Healthcheck status changed to failed")
	}
}

// HealthCheckRecovered is triggered when a health check recovers
func (sl *HealthCheckStatusListener) HealthCheckRecovered(status *health.State, recordedFailures int64, failureDurationSeconds float64) {
	statusJSON, err := json.Marshal(status)
	if err != nil {
		log.Error().Err(err).Msgf("Failed to parse health check status")
		log.Info().Msgf("Healthcheck status changed: Recovered from %d consecutive errors, lasting %1.2f seconds: %+v", recordedFailures, failureDurationSeconds, status)
	} else {
		log.Info().RawJSON("status", statusJSON).
			Str("recovered", fmt.Sprintf("Recovered from %d consecutive errors, lasting %1.2f seconds", recordedFailures, failureDurationSeconds)).
			Msg("Healthcheck status changed to recovered")
	}
}
