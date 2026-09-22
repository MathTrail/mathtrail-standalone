# What a person, or a deployment, needs to know after an apply.

output "public_url" {
  description = "The address the service calls itself, and the one the sign-in returns to."
  value       = local.public_url
}

output "service_url" {
  description = "The platform's own address for the service. Empty once it has been withdrawn."
  value       = google_cloud_run_v2_service.service.uri
}

output "domain_records" {
  description = "The DNS records the custom domain needs. Until they resolve, no certificate is issued and the domain answers nothing."
  value       = try(google_cloud_run_domain_mapping.service[0].status[0].resource_records, [])
}

output "image_repository" {
  description = "Where images are pushed; a deployed image is this, a slash, the service name, and a digest."
  value       = "${google_artifact_registry_repository.images.location}-docker.pkg.dev/${var.project_id}/${google_artifact_registry_repository.images.repository_id}"
}

output "runtime_service_account" {
  description = "The identity the service runs as."
  value       = google_service_account.runtime.email
}

output "deploy_service_account" {
  description = "The identity a deployment borrows."
  value       = google_service_account.deployer.email
}

output "workload_identity_provider" {
  description = "The provider a deployment presents its GitHub token to, in the full form the sign-in step expects."
  value       = google_iam_workload_identity_pool_provider.github.name
}
