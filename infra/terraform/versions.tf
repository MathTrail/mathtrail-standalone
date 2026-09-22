# What this configuration is pinned to, and where its state lives.
#
# The state bucket is not created here and cannot be: a backend has to exist
# before the configuration that would create it is read. It is made once by
# hand and named at init time, so one configuration can serve more than one
# deployment.

terraform {
  required_version = "1.16.3"

  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "8.3.0"
    }
  }

  backend "gcs" {}
}

provider "google" {
  project = var.project_id
  region  = var.region
}
