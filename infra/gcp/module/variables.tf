variable "project" {
  description = "The name of the project where the target clusters live."
  type        = string
}

variable "topic_name" {
  description = "The name of the topic interruption events will be published to"
  type        = string
}

variable "subscription_name" {
  description = "The name of the subscription that is subscribed to topic_name"
  type        = string
}

variable "log_sink_name" {
  description = "Name of the log sink"
  type = string
}

variable "log_sink_filter" {
  description = "Filter for the log sink"
  type = string
}

variable "labels" {
  description = "Labels to apply to all GCP resources created in this module"
  default = {
    "managed-by": "spot-interruption-exporter"
  }
}