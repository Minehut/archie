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

Using [KEDA PR #3564](https://github.com/kedacore/keda/pull/3564) fix for jetstream clustering 


## development

### running

Run `make run <args>`

### building

#### go releaser

Use `goreleaser` to build binaries, linux packages, and docker containers for all platforms without publishing.

Use `make release-snapshot` to build everything in `./dist/` and create local docker images.

#### go executable

Create a single platform binary in `./dist/` with `make build-go-linux-amd64` or `make build-go-linux-arm64` 
or `make build-go-local` to auto-detect the host platform.

#### docker images

Use `make docker-linux-amd64` or `make docker-linux-arm64` to build with local docker images.

## releasing

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

Set `release.disabled=false` in `.goreleaser.yaml`, and then run `make release`.

Go to the GitHub [Releases](https://github.com/superleaguegaming/archie/releases) page to publish the draft.

Restore `release.disabled=true`

### helm chart

#### install

Use `make helm-install` to add the nexus helm plugin and repo.

#### release

Update the `version` number in `helm/archie/Chart.yaml`.

Use `make helm-release` to publish the helm chart to the repo.
