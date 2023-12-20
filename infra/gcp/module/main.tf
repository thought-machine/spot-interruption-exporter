
resource "google_pubsub_topic" "topic" {
  name    = var.topic_name
  project = var.project

  message_retention_duration = "600s"
}

resource "google_logging_project_sink" "log_sink" {
  name    = var.log_sink_name
  project = var.project

  destination = "pubsub.googleapis.com/${google_pubsub_topic.topic.id}"
  filter      = var.log_sink_filter

  unique_writer_identity = true
}

resource "google_pubsub_topic_iam_binding" "binding" {
  project = var.project
  topic   = google_pubsub_topic.topic.name
  role    = "roles/pubsub.publisher"
  members = [
    google_logging_project_sink.log_sink.writer_identity,
  ]
}

resource "google_pubsub_subscription" "subscription" {
  name    = var.subscription_name
  topic   = google_pubsub_topic.topic.name
  project = var.project

  message_retention_duration = "600s"
  retain_acked_messages      = false

}

