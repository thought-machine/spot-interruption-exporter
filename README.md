# spot-interruption-exporter
Publishes a prometheus metric `interruption_events_total` that increments by 1 whenever a spot instance has been preempted.

This is a very helpful metric, as it 

- helps correlate workload issues with spot interruption times

- can aid in seeing if certain flavours are more susceptible to interruption

- can aid in seeing how much more susceptible single-zone clusters are to interruption

- can be used as a signal on whether to promote spot instances to other environments

The app can be expanded to support other cloud providers, but currently is only built for GCP.

A single deployment of the infrastructure and app is intended to serve all Kubernetes clusters in a given project.

## How it works
Spot preemption events are emitted as an audit log that contain the compute instance ID. These audit logs are forwarded to a pubsub topic via GCP Log Sink. The app then subscribes to this topic and handles the interruption event. 

The audit log for instance preemption does not contain information about the Kubernetes cluster the instance may or may not have been associated with. Since the node is already deleted by the time the preemption event is received, the compute API cannot be queried for more information. 

To work around this, the app keeps a mapping of compute instance ID to Kubernetes cluster. It can then use this when processing preemption events to publish the correct `kubernetes_cluster` label on the metric.

A second log router + pubsub topic exist to inform the app of new instances that belong to a Kubernetes cluster. On app startup, the compute API is queried to seed the mapping.

![spot-interruption-exporter-gcp](https://github.com/thought-machine/spot-interruption-exporter/assets/11613073/f2b01b81-1d13-4a2d-8303-9c842b51b3f7)

## Config

The app reads in a config file from `$CONFIG_PATH` with the structure below.

```yaml
log_level: debug
project_name: example-project
pubsub:
  instance_creation_subscription_name: sie-creation-subscription
  instance_interruption_subscription_name: sie-interruption-subscription
prometheus:
  port: 8090
  path: /metrics
```

## Deploying

### Infrastructure
You'll need to deploy the required infrastructure before standing up the application.

The infrastructure that the app depends for GCP on can be created via
```bash
$ terraform -chdir=infra/gcp init
$ terraform -chdir=infra/gcp apply
```

and can be destroyed via
```bash
$ terraform -chdir=infra/gcp destroy
```

### Kubernetes manifests
`kustomize/` holds relevant kubernetes config files. You will likely want to overlay the base resources. For an example of how you might do this, see `kustomize/example-overlay`.

## Verifying

You can send a test interruption message via
```bash
$ gcloud pubsub topics publish sie-interruption-topic --project <project> --message '{
  "protoPayload": {
    "@type": "type.googleapis.com/google.cloud.audit.AuditLog",
    "methodName": "compute.instances.preempted",
    "resourceName": "projects/mock-project/zones/europe-west1-c/instances/mock-instance-spot-3706-5b909138-nr65"
  }
}'
```

You can send a test instance creation message via 
```bash
$ gcloud pubsub topics publish sie-creation-topic --project <project> --message '{
  "protoPayload": {
    "@type": "type.googleapis.com/google.cloud.audit.AuditLog",
    "serviceName": "compute.googleapis.com",
    "methodName": "v1.compute.instances.insert",
    "resourceName": "projects/123456789/zones/europe-west1-c/instances/fake-resource",
    "request": {
      "labels": [
        {
          "key": "goog-k8s-cluster-name",
          "value": "fake-cluster"
        }
      ]
    }
  }
}'
```

After sending a few messages, you can view the metric count increasing
```bash
$ curl localhost:8080/metrics | grep interruption
# HELP interruption_events_total The total number of interruption events for a given cluster
# TYPE interruption_events_total counter
interruption_events_total{kubernetes_cluster="kubernetes-cluster"} 6
```
