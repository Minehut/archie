# archie

A worker to copy files from MinIO to any S3 Bucket using bucket event notifications and NATS JetStream

## usage

To configure use flags or env vars

## minio

Enable notifications to NATS JetStream in the MinIO server's environment.

```shell
MINIO_NOTIFY_NATS_ENABLE=on
MINIO_NOTIFY_NATS_TLS=on
MINIO_NOTIFY_NATS_JETSTREAM=on
MINIO_NOTIFY_NATS_ADDRESS=nats.archie.svc.cluster.local:4222
MINIO_NOTIFY_NATS_USERNAME=archie-pub
MINIO_NOTIFY_NATS_PASSWORD=xxxxx
MINIO_NOTIFY_NATS_SUBJECT=minioevents
MINIO_NOTIFY_NATS_QUEUE_DIR=/notify/nats
MINIO_NOTIFY_NATS_QUEUE_LIMIT=100000
```

## autoscaling

Use KEDA's [NATS JetStream Scaler](https://keda.sh/docs/latest/scalers/nats-jetstream/)

## development

Check out [DEVELOPER.md](DEVELOPER.md)

### app

#### install

```shell
brew install goreleaser
```

#### release

To use `goreleaser` set a new git tag:

```shell
git tag -a v0.1.0 -m "New release"
git push origin v0.1.0
```

Set a `GITHUB_TOKEN` with `repo` access.

Run `make release`.

The draft release should be ready to be published on the GitHub [Releases](https://github.com/superleaguegaming/archie/releases) page.

### helm chart

#### install

Use `make helm-install` to add the nexus helm plugin and repo.

#### release

Increment the chart `version` number in `helm/archie/Chart.yaml` and update the `appVersion` to the latest.

Use `make helm-release` to publish the helm chart to the repo.
