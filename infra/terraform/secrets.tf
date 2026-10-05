# The secrets the service is given, and nothing about what is in them: the
# values are added outside this configuration, and neither it nor its state
# ever sees them.

resource "google_secret_manager_secret" "seal_key" {
  secret_id = "${var.service_name}-seal-key"

  replication {
    auto {}
  }

  depends_on = [google_project_service.enabled]
}

# The key the children are counted under in the log: a secret of its own,
# never rotated with the sealing key, since a name that changed in the middle of
# a month would count one child twice.
resource "google_secret_manager_secret" "learner_key" {
  secret_id = "${var.service_name}-learner-key"

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

# The sign-in of a directory's reviewers: the password they type, and the demo
# account's grant at Google that it renews. Both stay empty, and the sign-in
# off, until a version of each is added by hand and named in the variables.
resource "google_secret_manager_secret" "reviewer_password" {
  secret_id = "${var.service_name}-reviewer-password"

  replication {
    auto {}
  }

  depends_on = [google_project_service.enabled]
}

resource "google_secret_manager_secret" "reviewer_grant" {
  secret_id = "${var.service_name}-reviewer-grant"

  replication {
    auto {}
  }

  depends_on = [google_project_service.enabled]
}

# The runtime identity may read these secrets, and that is the whole of what it
# may do anywhere in this project.
resource "google_secret_manager_secret_iam_member" "runtime" {
  for_each = {
    seal_key             = google_secret_manager_secret.seal_key.secret_id
    learner_key          = google_secret_manager_secret.learner_key.secret_id
    google_client_secret = google_secret_manager_secret.google_client_secret.secret_id
    reviewer_password    = google_secret_manager_secret.reviewer_password.secret_id
    reviewer_grant       = google_secret_manager_secret.reviewer_grant.secret_id
  }

  secret_id = each.value
  role      = "roles/secretmanager.secretAccessor"
  member    = "serviceAccount:${google_service_account.runtime.email}"
}
