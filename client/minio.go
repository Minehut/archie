package client

import (
	"context"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"io"
	"os"
	"time"
)

type Minio struct {
	client *minio.Client
}

type MinioObject struct {
	Bucket string
	Path   string
	Reader io.Reader
}

func (m *Minio) New(ctx context.Context, name, bucket, endpoint string, creds Credentials, useSSL bool, p Params, logLevel zerolog.Level) context.CancelFunc {

	//minio.MaxRetry = 0

	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(creds.MinioAccessKey, creds.MinioSecretAccessKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		log.Fatal().Err(err).Msgf("Failed to setup %s client", name)
	}

	if logLevel == zerolog.TraceLevel {
		client.TraceOn(os.Stdout)
	}

	if p == (Params{}) {
		log.Info().Msgf("Setup %s client to %s", name, client.EndpointURL())
	} else {
		log.Info().Msgf("Setup %s client to %s with %d threads and %dMB part size", name, client.EndpointURL(), p.Threads, p.PartSize)
	}

	bucketExists, err := client.BucketExists(ctx, bucket)
	if err != nil {
		log.Fatal().Err(err).Msgf("Failed to check if %s bucket %s exists", name, bucket)
	}

	if !bucketExists {
		log.Fatal().Msgf("%s bucket %s does not exist or access is missing", name, bucket)
	}

	// enable health checking
	healthCheckCancel, err := client.HealthCheck(5 * time.Second)
	if err != nil {
		log.Fatal().Msgf("Failed to start %s client health check", name)
	}

	m.client = client

	return healthCheckCancel
}

func (m *Minio) GetObject(ctx context.Context, bucket string, key string) (Object, error) {
	obj, err := m.client.GetObject(ctx, bucket, key, minio.GetObjectOptions{})
	if err != nil {
		return nil, err
	}
	var mo Object = &MinioObject{Bucket: bucket, Path: key, Reader: obj}
	return mo, nil
}

func (m *Minio) PutObject(ctx context.Context, bucket string, key string, reader io.Reader, objectSize int64, opts PutOptions) (UploadInfo, error) {
	putOpts := minio.PutObjectOptions{
		ContentType:    opts.ContentType,
		NumThreads:     opts.NumThreads,
		PartSize:       opts.PartSize,
		SendContentMd5: true,
	}

	if opts.ETag != "" {
		putOpts.UserMetadata = map[string]string{
			"Minio-Etag": opts.ETag,
		}
	}
	_, err := m.client.PutObject(ctx, bucket, key, reader, objectSize, putOpts)
	if err != nil {
		return UploadInfo{}, err
	}
	return UploadInfo{}, nil
}

func (m *Minio) RemoveObject(ctx context.Context, bucket string, key string) error {
	err := m.client.RemoveObject(ctx, bucket, key, minio.RemoveObjectOptions{})
	if err != nil {
		return err
	}
	return nil
}

func (m *Minio) IsOffline() bool {
	return m.client.IsOffline()
}

func (m *Minio) EndpointURL() string {
	return m.client.EndpointURL().String()
}

func (o *MinioObject) Stat(ctx context.Context) (*ObjectInfo, error) {
	srcStat, err := o.Reader.(*minio.Object).Stat()
	if err != nil {
		return nil, err
	}

	userMeta := srcStat.UserMetadata

	return &ObjectInfo{Size: srcStat.Size, ContentType: srcStat.ContentType, ETag: userMeta["Minio-Etag"]}, nil
}

func (o *MinioObject) GetReader() io.Reader {
	return o.Reader
}
