/**
 * GCP Spot Interruption Notifications
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
 * module "spot_interruption_notifier" {
 *   source       = "terraform.external.thoughtmachine.io/internal/spot-interruption-handler"
 *   version      = "0.0.1"
 *                            
 *   project      = <target-cluster's-project>
 * }
 * ```
 *
 * Deployment
 * ----------
 *
 * Deploying this module will create the following resources:
 *   * `google_pubsub_topic.instance_preemption`
 *   * `google_pubsub_subscription.instance_preemption`
 *   * `google_logging_project_sink.preemption_logs`
**/

locals {
  service_account_member = "serviceAccount.${var.project}.svc.id.goog[${var.kubernetes_service_account_namespace}/${var.kubernetes_service_account_name}]"

}

resource "google_pubsub_topic" "instance_preemption" {
  name    = var.topic_name
  project = var.project

  message_retention_duration = "600s"
}


resource "google_logging_project_sink" "preemption_logs" {
  name    = var.log_sink_name
  project = var.project

  destination = "pubsub.googleapis.com/${google_pubsub_topic.instance_preemption.id}"
  filter      = "protoPayload.methodName=\"compute.instances.preempted\""

  unique_writer_identity = true
}


resource "google_pubsub_subscription" "instance_preemption" {
  name    = var.subscription_name
  topic   = google_pubsub_topic.instance_preemption.name
  project = var.project

  message_retention_duration = "600s"
  retain_acked_messages      = false

}

resource "google_service_account" "spot_interruption_notifier" {
  account_id   = var.service_account_id
  display_name = "Spot Interruption Notifier"
  project      = var.project
}


resource "google_service_account_iam_binding" "workload_identity_user" {
  service_account_id = google_service_account.spot_interruption_notifier.name
  role               = "roles/iam.workloadIdentityUser"
  project            = var.project

  members = [
    local.service_account_member,
  ]
}

resource "google_service_account_iam_binding" "pubsub_subsriber" {
  service_account_id = google_service_account.spot_interruption_notifier.name
  role               = "roles/pubsub.subscriber"
  project            = var.project

  members = [
    local.service_account_member,
  ]
}