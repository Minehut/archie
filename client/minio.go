package client

import (
	"context"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/rs/zerolog/log"
	"os"
	"time"
)

type Params struct {
	Threads  uint
	PartSize uint64
}

func MinIO(ctx context.Context, name, bucket, endpoint, accessKey, secretAccessKey *string, useSSL *bool, p Params, trace *bool) (*minio.Client, context.CancelFunc) {

	//minio.MaxRetry = 0

	client, err := minio.New(*endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(*accessKey, *secretAccessKey, ""),
		Secure: *useSSL,
	})
	if err != nil {
		log.Fatal().Err(err).Msgf("Failed to setup %s client", *name)
	}

	if *trace {
		client.TraceOn(os.Stdout)
	}

	if p == (Params{}) {
		log.Info().Msgf("Setup %s client to %s", *name, client.EndpointURL())
	} else {
		log.Info().Msgf("Setup %s client to %s with %d threads and %dMB part size", *name, client.EndpointURL(), p.Threads, p.PartSize)
	}

	bucketExists, err := client.BucketExists(ctx, *bucket)
	if err != nil {
		log.Fatal().Err(err).Msgf("Failed to check if %s bucket %s exists", *name, *bucket)
	}

	if !bucketExists {
		log.Fatal().Msgf("%s bucket %s does not exist or access is missing", *name, *bucket)
	}

	// enable health checking
	destHealthCheckCancel, err := client.HealthCheck(5 * time.Second)
	if err != nil {
		log.Fatal().Msgf("Failed to start %s client health check", *name)
	}

	return client, destHealthCheckCancel
}
