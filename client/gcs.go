package client

import (
	"cloud.google.com/go/storage"
	_ "cloud.google.com/go/storage"
	"context"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"google.golang.org/api/option"
	"io"
)

type GCS struct {
	client   *storage.Client
	buffer   *[]byte
	endpoint string
}

type GCSObject struct {
	Bucket string
	Path   string
	Reader io.Reader
	Client *storage.Client
}

func (g *GCS) New(ctx context.Context, name, bucket, endpoint string, creds Credentials, useSSL bool, p Params, logLevel zerolog.Level) context.CancelFunc {
	var clientOptions []option.ClientOption

	if endpoint != "" {
		clientOptions = append(clientOptions, option.WithEndpoint(endpoint))
		g.endpoint = endpoint
	}

	if creds.GoogleCredentials != "" {
		clientOptions = append(clientOptions, option.WithCredentialsJSON([]byte(creds.GoogleCredentials)))
	}

	client, err := storage.NewClient(ctx, clientOptions...)
	if err != nil {
		log.Fatal().Err(err).Msgf("Failed to setup %s client", name)
	}

	// TODO: turn on debug in gcs client
	//if logLevel == zerolog.TraceLevel

	if p == (Params{}) {
		log.Info().Msgf("Setup %s client to %s", name, "GCS")
	} else {
		log.Info().Msgf("Setup %s client to %s with %d threads and %dMB part size", name, "GCS", p.Threads, p.PartSize)
	}

	b := client.Bucket(bucket)
	_, err = b.Attrs(ctx)
	if err != nil {
		if err.Error() == "storage: bucket doesn't exist" {
			log.Fatal().Msgf("%s bucket %s does not exist or access is missing", name, bucket)
		} else {
			log.Fatal().Err(err).Msgf("Failed to check if %s bucket %s exists", name, bucket)
		}
	}

	// the gcs client doesn't offer a health-check
	_, healthCheckCancel := context.WithCancel(ctx)

	g.client = client

	buffer := make([]byte, 100*1024*1024) // 100 MB
	g.buffer = &buffer

	return healthCheckCancel
}

func (g *GCS) GetObject(ctx context.Context, bucket string, key string) (Object, error) {
	obj, err := g.client.Bucket(bucket).Object(key).NewReader(ctx)
	if err != nil {
		return nil, err
	}
	var mo Object = &GCSObject{Bucket: bucket, Path: key, Reader: obj, Client: g.client}
	return mo, nil
}

func (g *GCS) PutObject(ctx context.Context, bucket string, key string, reader io.Reader, objectSize int64, opts PutOptions) (UploadInfo, error) {
	writer := g.client.Bucket(bucket).Object(key).NewWriter(ctx)
	writer.ChunkSize = int(opts.PartSize)
	writer.ContentType = opts.ContentType
	writer.Size = objectSize

	_, err := io.CopyBuffer(writer, reader, *g.buffer)
	if err != nil {
		return UploadInfo{}, err
	}

	err = writer.Close()
	if err != nil {
		return UploadInfo{}, err
	}

	return UploadInfo{}, nil
}

func (g *GCS) RemoveObject(ctx context.Context, bucket string, key string) error {
	if err := g.client.Bucket(bucket).Object(key).Delete(ctx); err != nil {
		if err != nil {
			return err
		}
	}
	return nil
}

func (g *GCS) IsOffline() bool {
	// the gcs library doesn't offer a health-check
	return false
}

func (g *GCS) EndpointURL() string {
	return g.endpoint
}

func (o *GCSObject) Stat(ctx context.Context) (*ObjectInfo, error) {
	obj, err := o.Client.Bucket(o.Bucket).Object(o.Path).Attrs(ctx)
	if err != nil {
		return nil, err
	}
	return &ObjectInfo{Size: obj.Size, ContentType: obj.ContentType}, nil
}

func (o *GCSObject) GetReader() io.Reader {
	return o.Reader
}
