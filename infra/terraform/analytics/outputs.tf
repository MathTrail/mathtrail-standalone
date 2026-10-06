# What a workflow outside the deployment needs of the counts.

output "snapshot_reader_email" {
  description = "The identity that reads the snapshot of the live numbers, and nothing else."
  value       = google_service_account.snapshot_reader.email
}
