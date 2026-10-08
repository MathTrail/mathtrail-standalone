# The daily report: a page of the owner's own report that Data Studio mails
# every morning, reading one view that sets what each day cost beside what the
# night counted of it.
#
# The costs are Cloud Billing's standard usage cost export, which only the
# console turns on, into the dataset below. The dataset is kept in the region
# the counts are kept in, so that one query reads both; a dataset in a region
# is given the costs from the day the export is turned on, and none before it.
# Turning the export on lets the export's own identity write here, and that
# right is the console's to keep: nothing here names who may read or write the
# dataset, so no apply takes it away. Whatever the export has written cannot
# be written again, so the dataset is not emptied for a destroy.

resource "google_bigquery_dataset" "billing" {
  dataset_id  = "billing"
  location    = var.region
  description = "Cloud Billing's standard usage cost export, which the console writes here once it is turned on."

  depends_on = [google_project_service.analytics]
}

# A view cannot be made over a table that is not there, and the export makes
# its table only once it has been turned on; the view waits for the deployment
# to say so.
resource "google_bigquery_table" "daily_report" {
  count = var.billing_export ? 1 : 0

  dataset_id          = google_bigquery_dataset.private.dataset_id
  table_id            = "daily_report"
  deletion_protection = false

  view {
    query = templatefile("${local.analytics}/billing/daily_report.sql", merge(local.names, {
      billing    = "${var.project_id}.${google_bigquery_dataset.billing.dataset_id}"
      project_id = var.project_id
    }))
    use_legacy_sql = false
  }

  depends_on = [google_bigquery_table.counts]
}
