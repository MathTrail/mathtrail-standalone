# How a deployment gets in. A pool trusts the tokens GitHub issues to one
# repository, and an identity with just enough rights to push an image and roll
# a revision borrows them. No service account key exists anywhere, so there is
# none to leak.

resource "google_iam_workload_identity_pool" "github" {
  workload_identity_pool_id = "${var.service_name}-github"
  display_name              = "GitHub Actions"
  description               = "Tokens GitHub Actions issues to the repository of this service"

  depends_on = [google_project_service.enabled]
}

resource "google_iam_workload_identity_pool_provider" "github" {
  workload_identity_pool_id          = google_iam_workload_identity_pool.github.workload_identity_pool_id
  workload_identity_pool_provider_id = "github"
  display_name                       = "GitHub Actions"

  # Both halves matter. The name says which repository, and the numeric owner id
  # says that it is still ours: an account that is renamed or deleted frees its
  # name for anyone to take, and the number goes with it.
  attribute_condition = "assertion.repository == \"${var.github_repository}\" && assertion.repository_owner_id == \"${var.github_owner_id}\""

  attribute_mapping = {
    "google.subject"       = "assertion.sub"
    "attribute.repository" = "assertion.repository"
  }

  oidc {
    issuer_uri = "https://token.actions.githubusercontent.com"
  }
}

resource "google_service_account" "deployer" {
  account_id   = "${var.service_name}-deploy"
  display_name = "Deployment identity of the ${var.service_name} service"

  depends_on = [google_project_service.enabled]
}

resource "google_service_account_iam_member" "deployer_from_github" {
  service_account_id = google_service_account.deployer.name
  role               = "roles/iam.workloadIdentityUser"
  member             = "principalSet://iam.googleapis.com/${google_iam_workload_identity_pool.github.name}/attribute.repository/${var.github_repository}"
}

# Push an image into the one repository,
resource "google_artifact_registry_repository_iam_member" "deployer" {
  location   = google_artifact_registry_repository.images.location
  repository = google_artifact_registry_repository.images.name
  role       = "roles/artifactregistry.writer"
  member     = "serviceAccount:${google_service_account.deployer.email}"
}

# roll a revision of the one service,
resource "google_cloud_run_v2_service_iam_member" "deployer" {
  location = google_cloud_run_v2_service.service.location
  name     = google_cloud_run_v2_service.service.name
  role     = "roles/run.developer"
  member   = "serviceAccount:${google_service_account.deployer.email}"
}

# and let that revision run as the runtime identity. Nothing wider: it cannot
# create services, change who may call them, or read a secret.
resource "google_service_account_iam_member" "deployer_runs_as_runtime" {
  service_account_id = google_service_account.runtime.name
  role               = "roles/iam.serviceAccountUser"
  member             = "serviceAccount:${google_service_account.deployer.email}"
}
