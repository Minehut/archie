package main

type Config struct {
	ApiVersion string `fig:"apiVersion" validate:"required"`

	LogLevel             string `fig:"logLevel" default:"info"`
	MsgTimeout           string `fig:"msgTimeout" default:"30m"`
	ShutdownWait         string `fig:"shutdownWait" default:"0s"`
	SkipLifecycleExpired bool   `fig:"skipLifecycleExpired"`

	Src struct {
		AccessKey string `fig:"accessKey"`
		Bucket    string `fig:"bucket"`
		Endpoint  string `fig:"endpoint" default:"localhost:9000"`
		Name      string `fig:"name" default:"destination"`
		SecretKey string `fig:"secretKey"`
		UseSSL    bool   `fig:"useSSL"`
	}

	Dest struct {
		AccessKey string `fig:"accessKey"`
		Bucket    string `fig:"bucket"`
		Endpoint  string `fig:"endpoint" default:"localhost:9000"`
		Name      string `fig:"name" default:"source"`
		PartSize  uint64 `fig:"partSize" default:"16"`
		SecretKey string `fig:"secretKey"`
		Threads   uint   `fig:"threads" default:"4"`
		UseSSL    bool   `fig:"useSSL"`
	}

	ExcludePaths struct {
		CopyObject   []string `fig:"copyObject"`
		RemoveObject []string `fig:"removeObject"`
	}

	HealthCheck struct {
		Disabled bool
		Port     int `default:"8080"`
	}

	Metrics struct {
		Port int `default:"9999"`
	}

	Jetstream struct {
		BatchSize            int    `fig:"batchSize" default:"1"`
		Password             string `fig:"password"`
		ProvisioningDisabled bool   `fig:"provisioningDisabled"`
		RootCA               string `fig:"rootCA"`
		Subject              string `fig:"subject" default:"archie-minio-events"`
		URL                  string `fig:"url" default:"nats://localhost:4222"`
		Username             string `fig:"username"`

		Stream struct {
			MaxAge           string `fig:"maxAge"`
			MaxSize          int64  `fig:"maxSize" default:"-1"`
			Name             string `fig:"name" default:"archie-stream"`
			Replicas         int    `fig:"replicas" default:"1"`
			RepublishSubject string `fig:"republishSubject"`
			Retention        string `fig:"retention" default:"limits"`
		}

		Consumer struct {
			Name          string `fig:"name" default:"archie-consumer"`
			MaxAckPending int    `fig:"maxAckPending" default:"1000"`
		}
	}
}

//jetStreamProvisioningDisabled := flag.Bool("jetstream-provisioning-disabled", LookupEnvOrBool("JETSTREAM_PROVISIONING_DISABLED", false), "disable the creation and configuration of the stream and consumer")
//jetStreamSubject := flag.String("jetstream-subject", LookupEnvOrString("JETSTREAM_SUBJECT", "archie-minio-events"), "nats jetstream subject to subscribe to")
//jetStreamURL := flag.String("jetstream-url", LookupEnvOrString("JETSTREAM_URL", "nats://localhost:4222"), "jetstream client url")
//jetStreamUsername := flag.String("jetstream-username", LookupEnvOrString("JETSTREAM_USERNAME", ""), "jetstream client username")
//metricsPort := flag.Int("metrics-port", LookupEnvOrInt("METRICS_PORT", 9999), "metrics tcp port number")
//msgTimeout := flag.String("msg-timeout", LookupEnvOrString("MSG_TIMEOUT", "30m"), "the max duration for a transfer, jetstream stream message ack timeout and internal transfer context timeout (set to -5s of this setting)")
//shutdownWait := flag.String("shutdown-wait", LookupEnvOrString("SHUTDOWN_WAIT", "0s"), "time to wait for running transfers to complete before exiting")
//skipLifecycleExpired := flag.Bool("skip-lifecycle-expired", LookupEnvOrBool("SKIP_LIFECYCLE_EXPIRED", false), "don't propagate deletes initiated by the lifecycle expiration")
//srcAccessKey := flag.String("src-access-key", LookupEnvOrString("SRC_ACCESS_KEY", ""), "source bucket access key")
//srcBucket := flag.String("src-bucket", LookupEnvOrString("SRC_BUCKET", "test"), "source bucket name")
//srcEndpoint := flag.String("src-endpoint", LookupEnvOrString("SRC_ENDPOINT", "localhost:9000"), "source endpoint")
//srcName := flag.String("src-name", LookupEnvOrString("SRC_NAME", "minio"), "source display name")
//srcSecretAccessKey := flag.String("src-secret-access-key", LookupEnvOrString("SRC_SECRET_ACCESS_KEY", ""), "source secret access key")
//srcUseSSL := flag.Bool("src-use-ssl", LookupEnvOrBool("SRC_USE_SSL", true), "use ssl for the source bucket")
//destAccessKey := flag.String("dest-access-key", LookupEnvOrString("DEST_ACCESS_KEY", ""), "destination bucket access key")
//destBucket := flag.String("dest-bucket", LookupEnvOrString("DEST_BUCKET", ""), "destination bucket name")
//destEndpoint := flag.String("dest-endpoint", LookupEnvOrString("DEST_ENDPOINT", "localhost:9000"), "destination endpoint")
//destName := flag.String("dest-name", LookupEnvOrString("DEST_NAME", "b2"), "destination display name")
//destPartSize := flag.Uint64("dest-part-size", LookupEnvOrUint64("DEST_PART_SIZE", 16), "upload part size in mebibytes")
//destSecretAccessKey := flag.String("dest-secret-access-key", LookupEnvOrString("DEST_SECRET_ACCESS_KEY", ""), "destination secret access key")
//destThreads := flag.Uint("dest-threads", LookupEnvOrUint("DEST_THREADS", 4), "number of upload threads")
//destUseSSL := flag.Bool("dest-use-ssl", LookupEnvOrBool("DEST_USE_SSL", true), "use ssl connection for the destination bucket")
//healthCheckEnabled := flag.Bool("health-check-enabled", LookupEnvOrBool("HEALTH_CHECK_ENABLED", true), "enable health-check server for k9s")
//healthCheckPort := flag.Int("health-check-port", LookupEnvOrInt("HEALTH_CHECK_PORT", 8080), "health check tcp port number")
//jetStreamBatchSize := flag.Int("jetstream-batch-size", LookupEnvOrInt("JETSTREAM_BATCH_SIZE", 1), "number of JetStream messages to pull per batch")
//jetStreamDurableConsumer := flag.String("jetstream-durable-consumer", LookupEnvOrString("JETSTREAM_DURABLE_CONSUMER", "archie-consumer"), "name of the durable stream consumer (queue group)")
//jetStreamMaxAckPending := flag.Int("jetstream-max-ack-pending", LookupEnvOrInt("JETSTREAM_MAX_ACK_PENDING", 1_000), "jetstream server will stop offering msgs for processing once it is waiting on too many un-ack'd msgs")
//jetStreamPassword := flag.String("jetstream-password", LookupEnvOrString("JETSTREAM_PASSWORD", ""), "jetstream client password")
//jetStreamRootCA := flag.String("jetstream-root-ca", LookupEnvOrString("JETSTREAM_ROOT_CA", ""), "path to the root CA cert file")
//jetStreamStream := flag.String("jetstream-stream", LookupEnvOrString("JETSTREAM_STREAM", "archie-stream"), "jetstream stream name")
//jetStreamStreamMaxAge := flag.String("jetstream-stream-max-age", LookupEnvOrString("JETSTREAM_STREAM_MAX_AGE", ""), "max duration to persist JetStream messages in the stream")
//jetStreamStreamMaxSize := flag.Int64("jetstream-stream-max-size", LookupEnvOrInt64("JETSTREAM_STREAM_MAX_SIZE", -1), "max size of stream in megabytes")
//jetStreamStreamReplicas := flag.Int("jetstream-stream-replicas", LookupEnvOrInt("JETSTREAM_STREAM_REPLICAS", 1), "number of replicas for the stream data")
//jetStreamStreamRetention := flag.String("jetstream-stream-retention", LookupEnvOrString("JETSTREAM_STREAM_RETENTION", "limits"), "stream retention policy: 'limits', 'interest', or 'work_queue'")
//jetStreamStreamRepublishSubject := flag.String("jetstream-stream-republish-subject", LookupEnvOrString("JETSTREAM_STREAM_REPUBLISH_SUBJECT", ""), "re-publish messages from the main subject to a separate subject")
