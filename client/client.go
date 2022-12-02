package client

import (
	"context"
	"github.com/rs/zerolog"
	"io"
)

type Client interface {
	EndpointURL() string
	GetObject(ctx context.Context, bucket string, key string) (Object, error)
	IsOffline() bool
	New(ctx context.Context, name, bucket, endpoint string, creds Credentials, useSSL bool, p Params, logLevel zerolog.Level) context.CancelFunc
	PutObject(ctx context.Context, bucket string, key string, reader io.Reader, objectSize int64, opts PutOptions) (UploadInfo, error)
	RemoveObject(ctx context.Context, bucket string, key string) error
}

type Object interface {
	GetReader() io.Reader
	Stat(ctx context.Context) (*ObjectInfo, error)
}

type PutOptions struct {
	ContentType string
	NumThreads  uint
	PartSize    uint64
}

type ObjectInfo struct {
	ContentType string
	Size        int64
}

type UploadInfo struct{}

type Credentials struct {
	MinioAccessKey       string
	MinioSecretAccessKey string
	GoogleCredentials    string
}

type Params struct {
	PartSize uint64
	Threads  uint
}
