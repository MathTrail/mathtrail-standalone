# The counts of how the service is used, kept for years for grant
# applications: a module of its own, on only where the deployment says so.
# A copy turns it on once its own privacy policy says what it keeps.

module "analytics" {
  source = "./analytics"
  count  = var.analytics ? 1 : 0

  project_id   = var.project_id
  region       = var.region
  service_name = var.service_name

  # Only a job GitHub runs in the environment live-numbers, which only main may
  # run in, borrows the snapshot's reader: the pool maps a token's subject, and
  # a job's subject names the environment it runs in.
  snapshot_reader = "principal://iam.googleapis.com/projects/${data.google_project.this.number}/locations/global/workloadIdentityPools/${var.workload_identity_pool_id}/subject/${var.github_oidc_subject_prefix}:environment:live-numbers"

  billing_export = var.billing_export

  depends_on = [google_project_service.enabled]
}
