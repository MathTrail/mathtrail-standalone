# The counts of how the service is used, kept for years for grant
# applications: a module of its own, on only where the deployment says so.
# A copy turns it on once its own privacy policy says what it keeps.

module "analytics" {
  source = "./analytics"
  count  = var.analytics ? 1 : 0

  project_id   = var.project_id
  region       = var.region
  service_name = var.service_name

  depends_on = [google_project_service.enabled]
}
