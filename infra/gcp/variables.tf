variable "project" {
  description = "The name of the project where the target cluster lives."
  type        = string
}

variable "topic_name" {
  description = "The name of the topic events will be published to"
  type        = string
  default     = "spot-interruption-exporter-topic"
}

variable "subscription_name" {
  description = "The name of the subscription that is subscribed to the topic where events are published"
  type        = string
  default     = "spot-interruption-exporter-subscription"
}

variable "log_sink_name" {
  description = "The name of the log sink that will sink events to the topic"
  type        = string
  default     = "spot-interruption-exporter-log-sink"
}

variable "service_account_id" {
  type        = string
  default     = "spot-interruption-notifier"
  description = "The account ID that is used for the service account's email address & unique ID, must be unique within a project"
}

variable "kubernetes_service_account_name" {
  type        = string
  default     = "spot-interruption-notifier-sa"
  description = "Name of the Kubernetes service account that will be bound to the spot-interruption-notifier pod. Will be used for workload identity."
}

variable "kubernetes_service_account_namespace" {
  type        = string
  default     = "spot-interruption-notifier"
  description = "Namespace of the Kubernetes service account that will be bound to the spot-interruption-notifier pod. Will be used for workload identity."
}