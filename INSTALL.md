# install

This guide will assist is with the setup of archie in kubernetes including the dependent systems (MinIO, NATS JetStream, and KEDA).

## nats

Create a `Certificate` as `nats-server-tls` for the NATS cluster and clients using instructions from [here](https://docs.nats.io/running-a-nats-service/nats-kubernetes/nats-cluster-and-cert-manager)
and [here](https://docs.nats.io/running-a-nats-service/configuration/securing_nats/tls#creating-self-signed-certificates-for-testing).
Note: the certificate must include the `usages` they include in the example.

```shell
➜ helm install --namespace archie --create-namespace --name nats nats-io/nats -f - <<-EOF
---
auth:
  basic:
    users:
    - password: rrrrr
      user: root
    - password: ppppp
      permissions:
        publish:
        - archie-minio-events
        subscribe:
        - _INBOX.>
      user: archie-pub
    - password: sssss
      permissions:
        allow_responses: true
        publish:
        - $JS.ACK.archie-stream.archie-consumer.*.*.*.*.*
        - $JS.API.INFO
        - $JS.API.STREAM.NAMES
        - $JS.API.STREAM.CREATE.archie-stream
        - $JS.API.STREAM.CREATE.archie-stream-archive
        - $JS.API.STREAM.INFO.archie-stream
        - $JS.API.STREAM.INFO.archie-stream-archive
        - $JS.API.STREAM.UPDATE.archie-stream
        - $JS.API.STREAM.UPDATE.archie-stream-archive
        - $JS.API.CONSUMER.CREATE.archie-stream.archie-consumer.archie-minio-events
        - $JS.API.CONSUMER.CREATE.archie-stream-archive.archie-consumer.archie-minio-events
        - $JS.API.CONSUMER.DURABLE.CREATE.archie-stream.archie-consumer
        - $JS.API.CONSUMER.DURABLE.CREATE.archie-stream-archive.archie-consumer
        - $JS.API.CONSUMER.INFO.archie-stream.archie-consumer
        - $JS.API.CONSUMER.INFO.archie-stream-archive.archie-consumer
        - $JS.API.CONSUMER.MSG.NEXT.archie-stream.archie-consumer
        - $JS.API.CONSUMER.MSG.NEXT.archie-stream-archive.archie-consumer
        subscribe:
        - _INBOX.>
        - archie-minio-events
      user: archie-sub
  enabled: true
cluster:
  enabled: true
  replicas: 3
  tls:
    cert: tls.crt
    key: tls.key
    secret:
      name: nats-server-tls
#exporter:
#  serviceMonitor:
#    enabled: true
#fileStorage:
#   storageClassName: test
k8sClusterDomain: cluster.local
nats:
  jetstream:
    enabled: true
    fileStorage:
      enabled: true
      size: 10Gi
      # storageClassName: test
    memStorage:
      enabled: false
  logging:
    debug: true
  tls:
    cert: tls.crt
    key: tls.key
    secret:
      name: nats-server-tls
    verify: false
EOF
```

## minio

Configure MinIO to send NATS notifications to the NATS JetStream cluster.

```shell
➜ helm upgrade --install \
  --namespace minio \
  --set mode=standalone \
  --set replicas=1 \
  --set rootUser="aaaaa" \
  --set rootPassword="bbbbb" \
  --set persistence.size=10Gi \
  --set resources.requests.memory=2Gi \
  --set trustedCertsSecret=minio-trusted-certs \
  --set environment.MINIO_NOTIFY_NATS_ENABLE=on \
  --set environment.MINIO_NOTIFY_NATS_JETSTREAM=on \
  --set environment.MINIO_NOTIFY_NATS_TLS=on \
  --set environment.MINIO_NOTIFY_NATS_TLS_SKIP_VERIFY=off \
  --set environment.MINIO_NOTIFY_NATS_ADDRESS="nats.nats.svc.cluster.local:4222" \
  --set environment.MINIO_NOTIFY_NATS_SUBJECT=archie-minio-events \
  --set environment.MINIO_NOTIFY_NATS_USERNAME=archie-pub \
  --set environment.MINIO_NOTIFY_NATS_PASSWORD=xxxxx \
  --set environment.MINIO_NOTIFY_NATS_QUEUE_DIR=/notify/nats \
  --set environment.MINIO_NOTIFY_NATS_QUEUE_LIMIT=100000 \
  minio \
  minio/minio
```

Look for this message on startup: `SQS ARNs: arn:minio:sqs::_:nats`

Create a bucket and enable notifications to NATS.

```shell
➜ mc alias set minio http://localhost:9000 "aaaaa" "bbbbb" --api s3v4
➜ mc mb minio/<bucket>
➜ mc event add minio/<bucket> "arn:minio:sqs::_:nats" --event put,delete
```

## keda

KEDA creates an HPA (horizontal pod autoscaler) that will autoscale the archie deployment based on the number of 'pending' messages in the queue.

```shell
➜ helm repo add kedacore https://kedacore.github.io/charts
➜ helm install --namespace keda --create-namespace --name keda kedacore/keda
```

## archie

Archie will provision and enforce configuration on the NATS JetStream stream and the consumer unless `.Values.jetstream.provisioningDisabled=true` is set.

```shell
➜ helm repo add superleaguegaming https://packages.slgg.io/repository/helm-hosted
➜ helm install --namespace archie --create-namespace --name archie superleaguegaming/archie -f - <<-EOF
---
jetstream:
  durableConsumer: archie-consumer
  stream: archie-stream
  streamReplicas: 3
  subject: archie-minio-events
  username: archie-sub
  password: sub-password
  metricsURL: nats.archie.svc.cluster.local:8222
  url: tls://nats.archie.svc.cluster.local:4222
keda:
  create: true
  maxReplicaCount: 50
  minReplicaCount: 5
  trigger:
    lagThreshold: 6
source:
  accessKey: xxxx
  bucket: source-bucket
  endpoint: source.endpoint
  name: source # just a label
  secretAccessKey: xxxx
  useSSL: false
destination:
  accessKey: xxxx
  bucket: destination-bucket
  endpoint: destination.endpoint
  name: destination # just a label
  secretAccessKey: xxxx
```
