# The module is configured by the root's provider, which pins the version.

terraform {
  required_providers {
    google = {
      source = "hashicorp/google"
    }
  }
}
