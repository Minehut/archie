# Configure

The archie server has a few cli options but requires a configuration file to operate.

## CLI Options

```shell
➜ archie <args>
  -config string
        config file path (default "config.yaml")
  -log-level string
        set the log level (default: info)
```

| Setting       | Description                       |
|---------------|-----------------------------------|
| `--config`    | config file path                  |
| `--log-level` | set the log level (default: info) |


## Config File Options

Combine each of the following sections to create a valid `config.yaml` file.

### Server Options

```yaml
apiVersion: v1

logLevel: info
shutdownWait: 30s
skipLifecycleExpired: true
msgTimeout: 15m
excludePaths:
  copyObject:
    - ^\w{3}/(\w+)/\1\.tar\.zst$
  removeObject:
    - ^\w{3}/(\w+)/\1\.tar\.zst$
```

| Setting                     | Description                                                                                                             |
|-----------------------------|-------------------------------------------------------------------------------------------------------------------------|
| `apiVersion`                | config file api version (required)                                                                                      |
| `logLevel`                  | set the log level (default: info)                                                                                       |
| `shutdownWait`              | time to wait for running transfers to complete before exiting                                                           |
| `skipLifecycleExpired`      | don't propagate deletes initiated by the minio lifecycle expiration                                                     |
| `msgTimeout`                | the max duration for a transfer includes the jetstream stream message ack timeout and internal transfer context timeout |
| `excludePaths.copyObject`   | list of paths as regex patterns to exclude from copy operations   (pcre support)                                        |
| `excludePaths.removeObject` | list of paths as regex patterns to exclude from delete operations (pcre support)                                        |


### JetStream Options

```yaml
jetstream:
  provisioningDisabled: false
  url: nats://nats.nats.svc.cluster.local:4222

  rootCA: /etc/nats-cert/nats-ca
  username: archie-sub
  password: abc123
  subject: archie-minio-events

  stream:
    name: archie-stream
    replicas: 3
    retention: interest
    maxSize: 2147
    maxAge: 720h

  consumer:
    name: nice1
    maxAckPending: 1000
    republishSubject: archie-minio-events-archive
```

| Flag                        | Description                                                                      |
|-----------------------------|----------------------------------------------------------------------------------|
| `batchSize`                 | number of messages the subscriber should fetch with each pull (default: 1)       |
| `username`                  | jetstream server username                                                        |
| `password`                  | jetstream server password                                                        |
| `subject`                   | nats subject for the pull subscriber                                             |
| `url`                       | url to the nats jetstream server (default: nats://localhost:4222)                |
| `rootCA`                    | path to the root CA file                                                         |
| `stream.name`               | stream name to use and/or create                                                 |
| `consumer.name`             | consumer name to use and/or create                                               |
| `provisioningDisabled`      | disable creation and configuration of the stream and consumer                    |
| `stream.replicas`           | stream to replicate the stream data                                              |
| `stream.retention`          | stream retention type to "limits", "interest", or "work-queue" (default: limits) |
| `stream.maxSize`            | stream max size in MB                                                            |
| `stream.maxAge`             | stream max age for messages using a go duration like "30m"                       |
| `consumer.maxAckPending`    | consumer max ack pending (default: 1000)                                         |
| `consumer.republishSubject` | consumer to re-publish messages to another subject                               |


### Transfer Source Options

```yaml
src:
  name: minio
  bucket: test
  endpoint: minio.minio.svc.cluster.local:9000
  useSSL: false
  accessKey: xxx
  secretKey: yyy
  googleCredentials: |
    {
      "type": "service_account",
      "project_id": "xxx",
      "private_key_id": "123"
      ...
    }
```

| Flag                | Description                                       |
|---------------------|---------------------------------------------------|
| `name`              | label name for the file source                    |
| `endpoint`          | endpoint (default: localhost:9000)                |
| `useSSL`            | enable ssl connection (default: false)            |
| `bucket`            | bucket name                                       |
| `accessKey`         | aws access key                                    |
| `secretKey`         | aws secret access key                             |
| `googleCredentials` | service account or refresh token JSON credentials |


### Transfer Destination Options

```yaml
dest:
  name: b2
  bucket: bucket-name
  endpoint: s3.us-west-004.somewhere.com
  useSSL: true
  threads: 4
  partSize: 16
  accessKey: xxx
  secretKey: yyy
  googleCredentials: |
    {
      "type": "service_account",
      "project_id": "xxx",
      "private_key_id": "123"
      ...
    }
```

| Flag                | Description                                       |
|---------------------|---------------------------------------------------|
| `name`              | label name for the file source                    |
| `endpoint`          | endpoint (default: localhost:9000)                |
| `useSSL`            | enable ssl connection (default: false)            |
| `bucket`            | bucket name                                       |
| `accessKey`         | aws access key                                    |
| `secretKey`         | aws secret access key                             |
| `threads`           | number of transfer threads (default: 4)           |
| `partSize`          | size of parts for uploads in MiB (default: 16)    |
| `googleCredentials` | service account or refresh token JSON credentials |

### Health Check Server Options

```yaml
healthCheck:
  disabled: false
  port: 8080
```

| Flag        | Description                 |
|-------------|-----------------------------|
| `disabled`  | disable health check server |
| `port`      | server listen port          |


### Metric Server Options

```yaml
metrics:
  port: 9999
```

| Flag        | Description                 |
|-------------|-----------------------------|
| `port`      | server listen port          |
