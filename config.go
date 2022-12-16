package main

type Config struct {
	ApiVersion string `fig:"apiVersion" validate:"required"`

	LogLevel             string `fig:"logLevel" default:"info"`
	MsgTimeout           string `fig:"msgTimeout" default:"30m"`
	ShutdownWait         string `fig:"shutdownWait" default:"0s"`
	SkipLifecycleExpired bool   `fig:"skipLifecycleExpired"`

	Src struct {
		AccessKey         string `fig:"accessKey"`
		Bucket            string `fig:"bucket"`
		Endpoint          string `fig:"endpoint"`
		GoogleCredentials string `fig:"googleCredentials"`
		Name              string `fig:"name" default:"destination"`
		SecretKey         string `fig:"secretKey"`
		UseSSL            bool   `fig:"useSSL"`
	}

	Dest struct {
		AccessKey         string `fig:"accessKey"`
		Bucket            string `fig:"bucket"`
		Endpoint          string `fig:"endpoint"`
		GoogleCredentials string `fig:"googleCredentials"`
		Name              string `fig:"name" default:"source"`
		PartSize          uint64 `fig:"partSize" default:"16"`
		SecretKey         string `fig:"secretKey"`
		Threads           uint   `fig:"threads" default:"4"`
		UseSSL            bool   `fig:"useSSL"`
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
