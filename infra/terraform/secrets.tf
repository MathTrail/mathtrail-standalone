# The two secrets the service is given, and nothing about what is in them: the
# values are added by hand, and neither this configuration nor its state ever
# sees them.

resource "google_secret_manager_secret" "seal_key" {
  secret_id = "${var.service_name}-seal-key"

  replication {
    auto {}
  }

  depends_on = [google_project_service.enabled]
}

resource "google_secret_manager_secret" "google_client_secret" {
  secret_id = "${var.service_name}-google-client-secret"

  replication {
    auto {}
  }

  depends_on = [google_project_service.enabled]
}

# The runtime identity may read these two secrets, and that is the whole of what
# it may do anywhere in this project.
resource "google_secret_manager_secret_iam_member" "runtime" {
  for_each = {
    seal_key             = google_secret_manager_secret.seal_key.secret_id
    google_client_secret = google_secret_manager_secret.google_client_secret.secret_id
  }

  secret_id = each.value
  role      = "roles/secretmanager.secretAccessor"
  member    = "serviceAccount:${google_service_account.runtime.email}"
}
