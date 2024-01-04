/**
 * GCP Spot Interruption Exporter
 * =============
 *
 * Description
 * -----------
 * This deploys a solution to have spot instance preemption events published to a pubsub topic
 * It requires cloudresourcemanager.googleapis.com to be enabled, along with the role `roles/logging.configWriter` being granted to the credentials used w terraform
 *
 * Usage
 * -----
 *
 * ```ts
 * module "spot_interruption_exporter" {
 *   source       = "infra/gcp"
 *
 *   project      = <target-cluster's-project>
 * }
 * ```
**/

locals {
  service_account_member = "serviceAccount:${var.project}.svc.id.goog[${var.kubernetes_service_account_namespace}/${var.kubernetes_service_account_name}]"
}

module "interruption_events" {
  source = "./event-forwarder"

  log_sink_filter   = "protoPayload.methodName=\"compute.instances.preempted\""
  log_sink_name     = "sie-interruption-sink"
  project           = var.project
  subscription_name = "sie-interruption-subscription"
  topic_name        = "sie-interruption-topic"
}

module "creation_events" {
  source = "./event-forwarder"

  log_sink_filter   = "protoPayload.serviceName=\"compute.googleapis.com\" AND protoPayload.methodName=\"v1.compute.instances.insert\" AND protoPayload.request.labels.key=\"goog-k8s-cluster-name\""
  log_sink_name     = "sie-creation-sink"
  project           = var.project
  subscription_name = "sie-creation-subscription"
  topic_name        = "sie-creation-topic"

}

resource "google_service_account" "spot_interruption_exporter" {
  account_id   = var.service_account_id
  display_name = "Spot Interruption Exporter"
  project      = var.project
}

resource "google_service_account_iam_binding" "workload_identity_user" {
  service_account_id = google_service_account.spot_interruption_exporter.name
  role               = "roles/iam.workloadIdentityUser"

  members = [
    local.service_account_member,
  ]
}

resource "google_project_iam_member" "pubsub_subscriber" {
  project = var.project
  role    = "roles/pubsub.subscriber"

  member = google_service_account.spot_interruption_exporter.member
}

resource "google_project_iam_member" "compute_read_only" {
  project = var.project
  role    = "roles/compute.viewer"

  member = google_service_account.spot_interruption_exporter.member
}
