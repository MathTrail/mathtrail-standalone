# What the counts need to know of the deployment around them.

variable "project_id" {
  description = "The project the service runs in, and the counts are kept in."
  type        = string
}

variable "region" {
  description = "Where the raw lines and the counts are kept: the service's own region."
  type        = string
}

variable "service_name" {
  description = "The name of the Cloud Run service whose lines are counted, and the prefix of the identity that counts them."
  type        = string
}

variable "snapshot_reader" {
  description = "The one federated principal that may borrow the identity reading the snapshot of the live numbers, as principal://iam.googleapis.com/projects/NUMBER/locations/global/workloadIdentityPools/POOL/subject/SUBJECT."
  type        = string
}

variable "billing_export" {
  description = "Whether Cloud Billing's standard usage cost export has made its table in the dataset billing, so that the daily report's view can be made over it."
  type        = bool
}
