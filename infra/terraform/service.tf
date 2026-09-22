# The service: the identity it runs as, the revision it serves, who may call it
# and the domain it answers on.

locals {
  public_url = "https://${var.public_host}"
}

# The service keeps no state and calls no Google API of its own — the child's
# profile is read with the parent's own credentials, not with this identity — so
# it exists in order to be able to do as little as possible. It reads two
# secrets; that is all it is ever granted.
resource "google_service_account" "runtime" {
  account_id   = "${var.service_name}-run"
  display_name = "Runtime identity of the ${var.service_name} service"

  depends_on = [google_project_service.enabled]
}

resource "google_cloud_run_v2_service" "service" {
  name     = var.service_name
  location = var.region

  # The chat hosts arrive over the public internet.
  ingress = "INGRESS_TRAFFIC_ALL"

  # One entrance, so that there is one issuer: once the platform's own address is
  # withdrawn, the custom domain is the only way in. The withdrawal comes after
  # the domain is mapped and answering, never before.
  default_uri_disabled = var.disable_default_url

  # What must never be lost is the project and the sign-in client, and neither
  # of them is described here; the service itself is rebuilt from this file in a
  # minute.
  deletion_protection = false

  template {
    service_account                  = google_service_account.runtime.email
    timeout                          = var.request_timeout
    max_instance_request_concurrency = var.concurrency

    # Down to nothing when nobody is asking, and up to a ceiling low enough that
    # a burst cannot outrun the free allowance.
    scaling {
      min_instance_count = 0
      max_instance_count = var.max_instances
    }

    containers {
      image = var.image

      resources {
        limits = {
          cpu    = var.cpu
          memory = var.memory
        }

        # Billed per request: an instance that is holding no request holds no CPU.
        cpu_idle = true
      }

      env {
        name  = "MATHTRAIL_PUBLIC_URL"
        value = local.public_url
      }

      env {
        name  = "MATHTRAIL_GOOGLE_CLIENT_ID"
        value = var.google_oauth_client_id
      }

      env {
        name = "MATHTRAIL_GOOGLE_CLIENT_SECRET"

        value_source {
          secret_key_ref {
            secret  = google_secret_manager_secret.google_client_secret.secret_id
            version = var.google_client_secret_version
          }
        }
      }

      env {
        name = "MATHTRAIL_SEAL_KEY_CURRENT"

        value_source {
          secret_key_ref {
            secret  = google_secret_manager_secret.seal_key.secret_id
            version = var.seal_key_version
          }
        }
      }

      # While a key is being rotated, what was sealed with the previous one still
      # has to open. It is an earlier version of the same secret, and outside a
      # rotation the service is given no previous key at all.
      dynamic "env" {
        for_each = var.seal_key_previous_version == "" ? [] : [var.seal_key_previous_version]

        content {
          name = "MATHTRAIL_SEAL_KEY_PREVIOUS"

          value_source {
            secret_key_ref {
              secret  = google_secret_manager_secret.seal_key.secret_id
              version = env.value
            }
          }
        }
      }

      dynamic "env" {
        for_each = var.settings

        content {
          name  = env.key
          value = env.value
        }
      }
    }
  }

  depends_on = [
    google_project_service.enabled,
    google_secret_manager_secret_iam_member.runtime,
  ]

  # The deployment owns which image is served, and this configuration owns
  # everything around it. Without this, the two would take turns: a deployment
  # would roll a new digest and the next apply would roll it straight back to
  # whatever the variable says. The variable is what a fresh project starts
  # with, and after that the field is not ours to set.
  lifecycle {
    ignore_changes = [template[0].containers[0].image]
  }
}

# Anyone may call it. The hosts arrive with no Google identity, and what decides
# whether a caller gets anything is the service's own token check.
resource "google_cloud_run_v2_service_iam_member" "public" {
  location = google_cloud_run_v2_service.service.location
  name     = google_cloud_run_v2_service.service.name
  role     = "roles/run.invoker"
  member   = "allUsers"
}

# The domain in front of the service. The DNS records it asks for have to be in
# place before its certificate can be issued, and creating it waits for that to
# finish.
resource "google_cloud_run_domain_mapping" "service" {
  count = var.create_domain_mapping ? 1 : 0

  location = var.region
  name     = var.public_host

  metadata {
    namespace = var.project_id
  }

  spec {
    route_name = google_cloud_run_v2_service.service.name
  }
}
