package main

import (
	"archie/archie"
	"archie/client"
	"context"
	"encoding/json"
	"flag"
	"github.com/kkyr/fig"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"go.arsenm.dev/pcre"
	"path/filepath"
	"sync"
	"time"
)

func main() {
	zerolog.TimeFieldFormat = time.RFC3339Nano

	configFile := flag.String("config", "config.yaml", "config file path")
	logLevelFlag := flag.String("log-level", LookupEnvOrString("LOG_LEVEL", ""), "set the log level (default: info)")
	flag.Parse()

	var cfg Config
	err := fig.Load(&cfg, fig.File(filepath.Base(*configFile)), fig.Dirs(".", filepath.Dir(*configFile)))
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load config file")
	}

	log.Info().
		Str("version", archie.Version).
		Str("releaseTag", archie.ReleaseTag).
		Str("commit", archie.ShortCommitID).
		Str("buildDate", archie.BuildDate).
		Msg("Starting archie")

	// compile and validate pcre regex exclude patterns
	var excludedPathCopyObject, excludedPathRemoveObject []*pcre.Regexp

	if len(cfg.ExcludePaths.CopyObject) > 0 || len(cfg.ExcludePaths.RemoveObject) > 0 {
		for _, excludedPathPattern := range cfg.ExcludePaths.CopyObject {
			excludedPathRegexp, err := pcre.Compile(excludedPathPattern)
			if err != nil {
				log.Fatal().Err(err).Str("pattern", excludedPathPattern).Msg("Failed to compile CopyObject pcre regex")
			}
			excludedPathCopyObject = append(excludedPathCopyObject, excludedPathRegexp)
		}

		for _, excludedPathPattern := range cfg.ExcludePaths.RemoveObject {
			excludedPathRegexp, err := pcre.Compile(excludedPathPattern)
			if err != nil {
				log.Fatal().Err(err).Str("pattern", excludedPathPattern).Msg("Failed to compile RemoveObject pcre regex")
			}
			excludedPathRemoveObject = append(excludedPathRemoveObject, excludedPathRegexp)
		}

		log.Info().Msgf("Regex patterns compiled with pcre v%s", pcre.Version())
	}

	var logLevel string
	// prefer cli arg over config
	if *logLevelFlag != "" {
		logLevel = *logLevelFlag
	} else {
		logLevel = cfg.LogLevel
	}

	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	if logLevel == "trace" {
		log.Info().Msg("Trace logging enabled")
		zerolog.SetGlobalLevel(zerolog.TraceLevel)
	} else if logLevel == "debug" {
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

	// log all config settings
	redactedCfg := cfg

	if redactedCfg.Src.AccessKey != "" {
		redactedCfg.Src.AccessKey = "REDACTED"
	}
	if redactedCfg.Src.SecretKey != "" {
		redactedCfg.Src.SecretKey = "REDACTED"
	}
	if redactedCfg.Src.GoogleCredentials != "" {
		redactedCfg.Src.GoogleCredentials = "REDACTED"
	}
	if redactedCfg.Dest.AccessKey != "" {
		redactedCfg.Dest.AccessKey = "REDACTED"
	}
	if redactedCfg.Dest.SecretKey != "" {
		redactedCfg.Dest.SecretKey = "REDACTED"
	}
	if redactedCfg.Dest.GoogleCredentials != "" {
		redactedCfg.Dest.GoogleCredentials = "REDACTED"
	}
	if redactedCfg.Jetstream.Password != "" {
		redactedCfg.Jetstream.Password = "REDACTED"
	}

	redactedCfgJSON, err := json.Marshal(redactedCfg)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to marshal config to json")
	}

	log.Info().RawJSON("cfg", redactedCfgJSON).Msg("Startup configuration")

	// archiver
	a := archie.Archiver{
		DestBucket:                cfg.Dest.Bucket,
		DestName:                  cfg.Dest.Name,
		DestPartSize:              cfg.Dest.PartSize,
		DestThreads:               cfg.Dest.Threads,
		FetchDone:                 make(chan string, 1),
		HealthCheckDisabled:       cfg.HealthCheck.Disabled,
		MaxRetries:                cfg.MaxRetries,
		MsgTimeout:                cfg.MsgTimeout,
		SkipEventBucketValidation: cfg.SkipEventBucketValidation,
		SkipLifecycleExpired:      cfg.SkipLifecycleExpired,
		SrcBucket:                 cfg.Src.Bucket,
		SrcName:                   cfg.Src.Name,
		WaitForMatchingETag:       cfg.WaitForMatchingETag,
		WaitGroup:                 &sync.WaitGroup{},
		ExcludePaths: struct {
			CopyObject   []*pcre.Regexp
			RemoveObject []*pcre.Regexp
		}{
			CopyObject:   excludedPathCopyObject,
			RemoveObject: excludedPathRemoveObject,
		},
	}

	var srcHealthCheckCancel, destHealthCheckCancel context.CancelFunc

	// source
	var c client.Client
	if cfg.Src.GoogleCredentials != "" {
		c = &client.GCS{}
	} else {
		c = &client.Minio{}
	}

	srcHealthCheckCancel = c.New(
		baseCtx,
		cfg.Src.Name,
		cfg.Src.Bucket,
		cfg.Src.Endpoint,
		client.Credentials{
			MinioSecretAccessKey: cfg.Src.SecretKey,
			MinioAccessKey:       cfg.Src.AccessKey,
			GoogleCredentials:    cfg.Src.GoogleCredentials,
		},
		cfg.Src.UseSSL,
		client.Params{},
		zerolog.GlobalLevel(),
	)

	a.SrcClient = c

	defer func() {
		log.Trace().Msg("Deferred source health check context canceled")
		srcHealthCheckCancel()
	}()

	// destination
	var d client.Client
	if cfg.Dest.GoogleCredentials != "" {
		d = &client.GCS{}
	} else {
		d = &client.Minio{}
	}

	destHealthCheckCancel = d.New(
		baseCtx,
		cfg.Dest.Name,
		cfg.Dest.Bucket,
		cfg.Dest.Endpoint,
		client.Credentials{
			MinioSecretAccessKey: cfg.Dest.SecretKey,
			MinioAccessKey:       cfg.Dest.AccessKey,
			GoogleCredentials:    cfg.Dest.GoogleCredentials,
		},
		cfg.Dest.UseSSL,
		client.Params{
			Threads:  cfg.Dest.Threads,
			PartSize: cfg.Dest.PartSize,
		},
		zerolog.GlobalLevel(),
	)

	a.DestClient = d

	defer func() {
		log.Trace().Msg("Deferred destination health check context canceled")
		destHealthCheckCancel()
	}()

	// queue
	jetStreamSub, jetStreamConn := client.JetStream(
		cfg.Jetstream.URL,
		cfg.Jetstream.Subject,
		cfg.Jetstream.Stream.Name,
		cfg.Jetstream.Consumer.Name,
		cfg.Jetstream.Stream.MaxAge,
		cfg.Jetstream.RootCA,
		cfg.Jetstream.Username,
		cfg.Jetstream.Password,
		cfg.Jetstream.Stream.Replicas,
		cfg.Jetstream.Consumer.MaxAckPending,
		cfg.Jetstream.Stream.MaxSize,
		cfg.MsgTimeout,
		cfg.Jetstream.Stream.Retention,
		cfg.Jetstream.Stream.RepublishSubject,
		cfg.Jetstream.ProvisioningDisabled,
	)

	// health check server
	healthCheckSrv := a.StartHealthCheckServer(cfg.HealthCheck.Port, jetStreamConn)

	// metrics server
	metricsSrv := a.StartMetricsServer(cfg.Metrics.Port)

	// single-thread message processor
	go a.MessageProcessor(baseCtx, msgCtx, jetStreamSub, cfg.Jetstream.BatchSize)

	// shutdown manager
	a.WaitForSignal(cfg.ShutdownWait, baseCancel, msgCancel, healthCheckSrv, metricsSrv)

	log.Info().Msg("Shutdown complete")
}
