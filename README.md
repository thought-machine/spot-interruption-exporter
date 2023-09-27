# spot-interruption-notifier
Publishes a prometheus metric `interruption_events_total` whenever a spot instance has been interrupted.

This is a very helpful metric, as it 

- helps correlate workload issues with spot interruption times

- can aid in seeing if certain flavours are more susceptible to interruption

- can aid in seeing how much more susceptible single-zone clusters are to interruption

- can be used as a signal on whether or not to promote spot instances to other environments

The app can be expanded to support other cloud providers, but currently is only built for GCP.

![spot-interruption-exporter-gcp](https://github.com/thought-machine/spot-interruption-exporter/assets/11613073/3aac4b50-8ff3-49b2-9edd-cf60da294e98)


## Config

The app reads in a config file from `$CONFIG_PATH` with the structure below.

```yaml
cloud_provider: gcp 
gcp:
  project_name: example
  subscription_name: spot-interruption-exporter-subscription 
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

You can send a test message via
```bash
$ gcloud pubsub topics publish spot-interruption-exporter-topic --message '{
  "protoPayload": {
    "@type": "type.googleapis.com/google.cloud.audit.AuditLog",
    "status": {
      "message": "Instance was preempted."
    },
    "authenticationInfo": {
      "principalEmail": "system@google.com"
    },
    "serviceName": "compute.googleapis.com",
    "methodName": "compute.instances.preempted",
    "resourceName": "projects/mock-project/zones/europe-west1-c/instances/mock-instance-spot-3706-5b909138-nr65",
    "request": {
      "@type": "type.googleapis.com/compute.instances.preempted"
    }
  },
  "insertId": "qnwer3e38dfz",
  "resource": {
    "type": "gce_instance",
    "labels": {
      "instance_id": "184448819...",
      "project_id": "mock-project",
      "zone": "europe-west1-c"
    }
  },
  "timestamp": "2023-09-16T10:42:31.325309Z",
  "severity": "INFO",
  "logName": "projects/mock-project/logs/cloudaudit.googleapis.com%2Fsystem_event",
  "operation": {
    "id": "systemevent-1694860946116....",
    "producer": "compute.instances.preempted",
    "first": true,
    "last": true
  },
  "receiveTimestamp": "2023-09-16T10:42:31.782066320Z"
}'
```

After sending a few messages, you can view the metric count increasing
```bash
$ curl localhost:8080/metrics | grep interruption
# HELP interruption_events_total The total number of interruption events for a given cluster
# TYPE interruption_events_total counter
interruption_events_total{kubernetes_cluster="kubernetes-cluster"} 6
```
