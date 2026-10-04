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
