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
