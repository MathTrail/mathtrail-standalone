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
  description = "Where images are pushed. The image of the service is this, a slash, and the name of the binary in it."
  value       = "${google_artifact_registry_repository.images.location}-docker.pkg.dev/${var.project_id}/${google_artifact_registry_repository.images.repository_id}"
}

output "service_name" {
  description = "The name a deployment names when it rolls a new revision."
  value       = google_cloud_run_v2_service.service.name
}

output "region" {
  description = "The region a deployment names alongside it."
  value       = google_cloud_run_v2_service.service.location
}

output "runtime_service_account" {
  description = "The identity the service runs as."
  value       = google_service_account.runtime.email
}

output "deploy_service_account" {
  description = "The identity a deployment borrows to push an image and roll a revision."
  value       = google_service_account.deployer.email
}

output "secret_seal_key" {
  description = "The secret holding the sealing key. It has to hold a version before a revision can start."
  value       = google_secret_manager_secret.seal_key.secret_id
}

output "secret_google_client" {
  description = "The secret holding the Google client secret, on the same condition."
  value       = google_secret_manager_secret.google_client_secret.secret_id
}
