# spot-interruption-notifier
Publishes a prometheus metric `interruption_events_total` whenever a spot instance has been interrupted.

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
`kustomize/` holds relevant kubernetes config files. You will likely want to overlay the base resources. For an example of how you might do this, see `kustomize/example-overlay`.

## GCP overview
![gcp_layout.png](img/gcp.png)

### Local development

The infrastructure that the app depends for GCP on can be created via
```bash
$ terraform -chdir=infra/gcp init
$ terraform -chdir=infra/gcp apply
```

and can be destroyed via
```bash
$ terraform -chdir=infra/gcp destroy
```

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

## AWS Overview
This has not been implemented for AWS, although something similar could be achieved using EventBridge -> SQS Queue.
