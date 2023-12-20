variable "project" {
  description = "The name of the project where the target clusters lives"
  type        = string
}

variable "service_account_id" {
  type        = string
  default     = "spot-interruption-exporter"
  description = "The account ID that is used for the service account's email address & unique ID, must be unique within a project"
}

variable "kubernetes_service_account_name" {
  type        = string
  default     = "spot-interruption-exporter-sa"
  description = "Name of the Kubernetes service account that will be bound to the spot-interruption-exporter pod. Will be used for workload identity."
}

variable "kubernetes_service_account_namespace" {
  type        = string
  default     = "spot-interruption-exporter"
  description = "Namespace of the Kubernetes service account that will be bound to the spot-interruption-exporter pod. Will be used for workload identity."
}