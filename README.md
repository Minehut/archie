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

## known issues

* KEDA needed a patch to fix the scaler for using jetstream in a cluster - [PR #3564](https://github.com/kedacore/keda/pull/3564)
* NATS-Exporter needed to pass the `first_seq` stream info - [PR #190](https://github.com/nats-io/prometheus-nats-exporter/pull/190)
* MinIO doesn't reconnect to NATS server if it is down for a while - Create PR
* The NATS JetStream stream's first sequence metric is unstable - Create PR
