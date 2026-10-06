# The snapshot of the live numbers the site's page "Research" shows: the one
# value live.sql gives of the public views, kept every day in a dataset of its
# own. The public views are authorized nowhere, so whoever reads them must be
# able to read the counts under them too. The snapshot's reader reads this
# table instead, as a table is browsed, which runs no query and needs no right
# but to read it.

resource "google_bigquery_dataset" "site" {
  dataset_id  = "impact_site"
  location    = var.region
  description = "The snapshot of the live numbers the site's page \"Research\" shows: the public views' latest month counted whole, one row, made again every day."

  # Its one table is the scheduled query's, made again by the query's next run.
  delete_contents_on_destroy = true

  depends_on = [google_project_service.analytics]
}

# The counter makes the snapshot: it reads the public views over the counts it
# writes already, and replaces the one table.
resource "google_bigquery_dataset_iam_member" "counter_reads_public" {
  dataset_id = google_bigquery_dataset.public.dataset_id
  role       = "roles/bigquery.dataViewer"
  member     = "serviceAccount:${google_service_account.counter.email}"
}

resource "google_bigquery_dataset_iam_member" "counter_writes_snapshot" {
  dataset_id = google_bigquery_dataset.site.dataset_id
  role       = "roles/bigquery.dataEditor"
  member     = "serviceAccount:${google_service_account.counter.email}"
}

# An hour after every night the snapshot is taken again, whole: its one row
# replaces the table, which the first run makes.
resource "google_bigquery_data_transfer_config" "snapshot" {
  display_name           = "${var.service_name} snapshot of the live numbers"
  location               = var.region
  data_source_id         = "scheduled_query"
  schedule               = "every day 02:30"
  service_account_name   = google_service_account.counter.email
  destination_dataset_id = google_bigquery_dataset.site.dataset_id

  params = {
    query                           = templatefile("${local.analytics}/site/live.sql", local.names)
    destination_table_name_template = "live"
    write_disposition               = "WRITE_TRUNCATE"
  }

  depends_on = [
    google_project_iam_member.counter_jobs,
    google_bigquery_dataset_iam_member.counter_writes_counts,
    google_bigquery_dataset_iam_member.counter_reads_public,
    google_bigquery_dataset_iam_member.counter_writes_snapshot,
    google_bigquery_table.closed_months,
    google_bigquery_table.public,
  ]
}

# Who reads the snapshot: no role on the project, the snapshot's dataset and
# nothing else, and borrowed by one federated identity alone.
resource "google_service_account" "snapshot_reader" {
  account_id   = "${var.service_name}-live"
  display_name = "Reads the snapshot of the live numbers of the ${var.service_name} site"

  depends_on = [google_project_service.analytics]
}

resource "google_bigquery_dataset_iam_member" "snapshot_reader_reads" {
  dataset_id = google_bigquery_dataset.site.dataset_id
  role       = "roles/bigquery.dataViewer"
  member     = "serviceAccount:${google_service_account.snapshot_reader.email}"
}

resource "google_service_account_iam_member" "snapshot_reader_borrowed" {
  service_account_id = google_service_account.snapshot_reader.name
  role               = "roles/iam.workloadIdentityUser"
  member             = var.snapshot_reader
}
