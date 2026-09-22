# How a deployment gets in: an identity with just enough rights to push an image
# and roll a revision, borrowed through the tokens GitHub issues to one
# repository. No service account key exists anywhere, so there is none to leak.
#
# The pool those tokens are exchanged in is deliberately not described here. It
# is what this configuration itself is applied through, and a door cannot be
# built by whoever walks in by it: the pool, its provider and the identity that
# runs Terraform are created once, from outside, and named here by id.

resource "google_service_account" "deployer" {
  account_id   = "${var.service_name}-deploy"
  display_name = "Deployment identity of the ${var.service_name} service"

  depends_on = [google_project_service.enabled]
}

resource "google_service_account_iam_member" "deployer_from_github" {
  service_account_id = google_service_account.deployer.name
  role               = "roles/iam.workloadIdentityUser"
  member             = "principalSet://iam.googleapis.com/projects/${data.google_project.this.number}/locations/global/workloadIdentityPools/${var.workload_identity_pool_id}/attribute.repository/${var.github_repository}"
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
