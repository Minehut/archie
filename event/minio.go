package event

import "time"

type Minio struct {
	EventName string   `json:"EventName"`
	Key       string   `json:"Key"`
	Records   []Record `json:"Records"`
}

type UserIdentity struct {
	PrincipalID string `json:"principalId"`
}

type RequestParameters struct {
	PrincipalID     string `json:"principalId"`
	Region          string `json:"region,omitempty"`
	SourceIPAddress string `json:"sourceIPAddress"`
}

type ResponseElements struct {
	ContentLength        string `json:"content-length"`
	XAmzRequestID        string `json:"x-amz-request-id"`
	XMinioDeploymentID   string `json:"x-minio-deployment-id"`
	XMinioOriginEndpoint string `json:"x-minio-origin-endpoint"`
}

type OwnerIdentity struct {
	PrincipalID string `json:"principalId"`
}

type Bucket struct {
	Name          string        `json:"name"`
	OwnerIdentity OwnerIdentity `json:"ownerIdentity"`
	Arn           string        `json:"arn"`
}

type UserMetadata struct {
	ContentType string `json:"content-type"`
}

type Object struct {
	Key          string       `json:"key"`
	Size         int64        `json:"size"`
	ETag         string       `json:"eTag"`
	ContentType  string       `json:"contentType"`
	UserMetadata UserMetadata `json:"userMetadata"`
	Sequencer    string       `json:"sequencer"`
}

type S3 struct {
	S3SchemaVersion string `json:"s3SchemaVersion"`
	ConfigurationID string `json:"configurationId"`
	Bucket          Bucket `json:"bucket"`
	Object          Object `json:"object"`
}

type Source struct {
	Host      string `json:"host"`
	Port      string `json:"port"`
	UserAgent string `json:"userAgent"`
}

type Record struct {
	EventVersion      string            `json:"eventVersion"`
	EventSource       string            `json:"eventSource"`
	AwsRegion         string            `json:"awsRegion,omitempty"`
	EventTime         time.Time         `json:"eventTime"`
	EventName         string            `json:"eventName"`
	UserIdentity      UserIdentity      `json:"userIdentity"`
	RequestParameters RequestParameters `json:"requestParameters"`
	ResponseElements  ResponseElements  `json:"responseElements"`
	S3                S3                `json:"s3"`
	Source            Source            `json:"source"`
}
