# archie

A scalable system to replicate files from MinIO buckets to any S3 compatible bucket using event notifications, NATS JetStream exactly-once delivery, and KEDA autoscaling in kubernetes.

app features:
* replicate bucket data with minio-go
  * copy and remove files
* async healthcheck server
* prometheus metrics server
* nats jetstream provisioning
* efficient queue pull consumer
* graceful shutdown wait timer
* ignore lifecycle expirations
* exclude paths with pcre regex

## install

Check out the helm chart [INSTALL.md](INSTALL.md)

chart features:
* archie worker deployment
* keda `ScaledObject` deployment scaler
* prometheus `ServiceMonitor` metrics
* prometheus `PrometheusRules` alerts

## queue

Use [NATS JetStream](https://docs.nats.io/nats-concepts/jetstream) to queue bucket event notifications from MinIO.

## autoscaling

Use KEDA's [NATS JetStream Scaler](https://keda.sh/docs/latest/scalers/nats-jetstream/) to scale the workers.

## development

Check out [DEVELOPER.md](DEVELOPER.md)

## known issues

* KEDA needed a patch to fix the scaler for using jetstream in a cluster - [PR #3564](https://github.com/kedacore/keda/pull/3564) (waiting)
* NATS-Exporter needed to pass the `first_seq` stream info - [PR #190](https://github.com/nats-io/prometheus-nats-exporter/pull/190) (merged)
* NATS JetStream stream's first sequence metric is unstable - TODO: Create PR
* MinIO doesn't reconnect to NATS server if it is down for a while - TODO: Create PR
* PCRE Regex module somewhat limits our build OS and ARCH - [INFO](https://gitea.arsenm.dev/Arsen6331/pcre#supported-goos-goarch)
