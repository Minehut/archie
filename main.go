package main

import (
	"archie/archie"
	"archie/client"
	"context"
	"flag"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"sync"
	"time"
)

func main() {
	//TODO: verify all settings are output somewhere
	debug := flag.Bool("debug", LookupEnvOrBool("DEBUG", false), "set log level to debug")
	destAccessKey := flag.String("dest-access-key", LookupEnvOrString("DEST_ACCESS_KEY", ""), "destination bucket access key")
	destBucket := flag.String("dest-bucket", LookupEnvOrString("DEST_BUCKET", ""), "destination bucket name")
	destEndpoint := flag.String("dest-endpoint", LookupEnvOrString("DEST_ENDPOINT", "localhost:9000"), "destination endpoint")
	destName := flag.String("dest-name", LookupEnvOrString("DEST_NAME", "b2"), "destination display name")
	destPartSize := flag.Uint64("dest-part-size", LookupEnvOrUint64("DEST_PART_SIZE", 16), "upload part size in mebibytes")
	destSecretAccessKey := flag.String("dest-secret-access-key", LookupEnvOrString("DEST_SECRET_ACCESS_KEY", ""), "destination secret access key")
	destThreads := flag.Uint("dest-threads", LookupEnvOrUint("DEST_THREADS", 4), "number of upload threads")
	destUseSSL := flag.Bool("dest-use-ssl", LookupEnvOrBool("DEST_USE_SSL", true), "use ssl connection for the destination bucket")
	healthCheckEnabled := flag.Bool("health-check-enabled", LookupEnvOrBool("HEALTH_CHECK_ENABLED", true), "enable health-check server for k9s")
	healthCheckPort := flag.Int("health-check-port", LookupEnvOrInt("HEALTH_CHECK_PORT", 8080), "health check tcp port number")
	jetStreamBatchSize := flag.Int("jetstream-batch-size", LookupEnvOrInt("JETSTREAM_BATCH_SIZE", 1), "number of JetStream messages to pull per batch")
	jetStreamDurableConsumer := flag.String("jetstream-durable-consumer", LookupEnvOrString("JETSTREAM_DURABLE_CONSUMER", "durable"), "name of the durable stream consumer (queue group)")
	jetStreamMaxAckPending := flag.Int("jetstream-max-ack-pending", LookupEnvOrInt("JETSTREAM_MAX_ACK_PENDING", 1_000), "jetstream server will stop offering msgs for processing once it is waiting on too many un-ack'd msgs")
	jetStreamPassword := flag.String("jetstream-password", LookupEnvOrString("JETSTREAM_PASSWORD", ""), "jetstream client password")
	jetStreamRootCA := flag.String("jetstream-root-ca", LookupEnvOrString("JETSTREAM_ROOT_CA", ""), "path to the root CA cert file")
	jetStreamStream := flag.String("jetstream-stream", LookupEnvOrString("JETSTREAM_STREAM", "archie-stream"), "jetstream stream name")
	jetStreamStreamMaxAge := flag.String("jetstream-stream-max-age", LookupEnvOrString("JETSTREAM_STREAM_MAX_AGE", "72h"), "max duration to persist JetStream messages in the stream")
	jetStreamStreamMaxSize := flag.Int64("jetstream-stream-max", LookupEnvOrInt64("JETSTREAM_STREAM_MAX_SIZE", -1), "max size of stream in megabytes")
	jetStreamStreamReplicas := flag.Int("jetstream-stream-replicas", LookupEnvOrInt("JETSTREAM_STREAM_REPLICAS", 1), "number of replicas for the stream data")
	jetreamStreamRePublishEnabled := flag.Bool("jetstream-stream-republish-enabled", LookupEnvOrBool("JETSTREAM_STREAM_REPUBLISH_ENABLED", false), "re-publish messages from the main stream to a separate archive stream")
	jetStreamSubject := flag.String("jetstream-subject", LookupEnvOrString("JETSTREAM_SUBJECT", "minioevents"), "nats jetstream subject to subscribe to")
	jetStreamURL := flag.String("jetstream-url", LookupEnvOrString("JETSTREAM_URL", "nats://localhost:4222"), "jetstream client url")
	jetStreamUsername := flag.String("jetstream-username", LookupEnvOrString("JETSTREAM_USERNAME", ""), "jetstream client username")
	metricsPort := flag.Int("metrics-port", LookupEnvOrInt("METRICS_PORT", 9999), "metrics tcp port number")
	msgTimeout := flag.String("msg-timeout", LookupEnvOrString("MSG_TIMEOUT", "30m"), "the max duration for a transfer, jetstream stream message ack timeout and internal transfer context timeout (set to -5s of this setting)")
	shutdownWait := flag.String("shutdown-wait", LookupEnvOrString("SHUTDOWN_WAIT", "0s"), "time to wait for running transfers to complete before exiting")
	skipLifecycleExpired := flag.Bool("skip-lifecycle-expired", LookupEnvOrBool("SKIP_LIFECYCLE_EXPIRED", false), "don't propagate deletes initiated by the lifecycle expiration")
	srcAccessKey := flag.String("src-access-key", LookupEnvOrString("SRC_ACCESS_KEY", ""), "source bucket access key")
	srcBucket := flag.String("src-bucket", LookupEnvOrString("SRC_BUCKET", "test"), "source bucket name")
	srcEndpoint := flag.String("src-endpoint", LookupEnvOrString("SRC_ENDPOINT", "localhost:9000"), "source endpoint")
	srcName := flag.String("src-name", LookupEnvOrString("SRC_NAME", "minio"), "source display name")
	srcSecretAccessKey := flag.String("src-secret-access-key", LookupEnvOrString("SRC_SECRET_ACCESS_KEY", ""), "source secret access key")
	srcUseSSL := flag.Bool("src-use-ssl", LookupEnvOrBool("SRC_USE_SSL", true), "use ssl for the source bucket")
	trace := flag.Bool("trace", LookupEnvOrBool("TRACE", false), "set log level to trace")

	flag.Parse()

	zerolog.TimeFieldFormat = time.RFC3339Nano
	log.Info().
		Str("version", archie.Version).
		Str("releaseTag", archie.ReleaseTag).
		Str("commit", archie.ShortCommitID).
		Str("buildDate", archie.BuildDate).
		Msg("Starting archie")

	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	if *trace {
		log.Info().Msg("Trace logging enabled")
		zerolog.SetGlobalLevel(zerolog.TraceLevel)
	} else if *debug {
		log.Info().Msg("Debug logging enabled")
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}

	// base context - cancel message processing (give time to let active transfers finish)
	baseCtx, baseCancel := context.WithCancel(context.Background())
	defer func() {
		log.Trace().Msg("Deferred base context canceled")
		baseCancel()
	}()

	// msg context - cancel active transfers (immediately)
	msgCtx, msgCancel := context.WithCancel(context.Background())
	defer func() {
		log.Trace().Msg("Deferred msg context canceled")
		msgCancel()
	}()

	// archiver
	a := archie.Archiver{
		DestBucket:           *destBucket,
		DestName:             *destName,
		DestPartSize:         *destPartSize,
		DestThreads:          *destThreads,
		FetchDone:            make(chan string, 1),
		HealthCheckEnabled:   *healthCheckEnabled,
		MsgTimeout:           *msgTimeout,
		SkipLifecycleExpired: *skipLifecycleExpired,
		SrcBucket:            *srcBucket,
		SrcName:              *srcName,
		WaitGroup:            &sync.WaitGroup{},
	}

	var srcHealthCheckCancel, destHealthCheckCancel context.CancelFunc

	// source
	a.SrcClient, srcHealthCheckCancel = client.MinIO(
		baseCtx,
		srcName,
		srcBucket,
		srcEndpoint,
		srcAccessKey,
		srcSecretAccessKey,
		srcUseSSL,
		client.Params{},
		trace,
	)

	defer func() {
		log.Trace().Msg("Deferred source health check context canceled")
		srcHealthCheckCancel()
	}()

	// destination
	a.DestClient, destHealthCheckCancel = client.MinIO(
		baseCtx,
		destName,
		destBucket,
		destEndpoint,
		destAccessKey,
		destSecretAccessKey,
		destUseSSL,
		client.Params{Threads: *destThreads, PartSize: *destPartSize},
		trace,
	)

	defer func() {
		log.Trace().Msg("Deferred destination health check context canceled")
		destHealthCheckCancel()
	}()

	// queue
	jetStreamSub, jetStreamConn := client.JetStream(
		*jetStreamURL,
		*jetStreamSubject,
		*jetStreamStream,
		*jetStreamDurableConsumer,
		*jetStreamStreamMaxAge,
		*jetStreamRootCA,
		*jetStreamUsername,
		*jetStreamPassword,
		*jetStreamStreamReplicas,
		*jetStreamMaxAckPending,
		*jetStreamStreamMaxSize,
		*msgTimeout,
		*jetreamStreamRePublishEnabled,
	)

	// health check server
	healthCheckSrv := a.StartHealthCheckServer(*healthCheckPort, jetStreamConn)

	// metrics server
	metricsSrv := a.StartMetricsServer(*metricsPort)

	// single-thread message processor
	go a.MessageProcessor(baseCtx, msgCtx, jetStreamSub, *jetStreamBatchSize)

	// shutdown manager
	a.WaitForSignal(*shutdownWait, baseCancel, msgCancel, healthCheckSrv, metricsSrv)

	log.Info().Msg("Shutdown complete")
}
