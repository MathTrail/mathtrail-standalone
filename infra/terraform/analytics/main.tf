# The counts of how the service is used, kept for years for grant
# applications, and nothing that could name a child kept beside them.
#
# The lines the children are counted from are copied into a log bucket of
# their own and kept there 62 days. Every night a query counts them again into
# tables that hold no name of any child, and those tables are what is kept.
# Two sets of views read the tables: exact numbers for the owner, and, for
# anyone, numbers that show no group of fewer than ten children. The SQL is in
# infra/analytics/, beside the list of events the lines are copied for.

locals {
  analytics = "${path.module}/../../analytics"

  # The events the children are counted from, as the service names their
  # lines.
  events = jsondecode(file("${local.analytics}/events.json"))

  # The names the SQL is written with, filled in by templatefile.
  names = {
    logs   = "${var.project_id}.${google_logging_linked_dataset.activity.link_id}._AllLogs"
    impact = "${var.project_id}.${google_bigquery_dataset.impact.dataset_id}"
    public = "${var.project_id}.${google_bigquery_dataset.public.dataset_id}"
    events = join(", ", [for event in local.events : "\"${event}\""])
  }

  tables         = toset([for file in fileset("${local.analytics}/tables", "*.json") : trimsuffix(file, ".json")])
  private_views  = toset([for file in fileset("${local.analytics}/views/private", "*.sql") : trimsuffix(file, ".sql")])
  public_parts   = toset(["fold", "closed_months", "learners_by"])
  public_views   = setsubtract([for file in fileset("${local.analytics}/views/public", "*.sql") : trimsuffix(file, ".sql")], local.public_parts)
  dimensions     = toset(["host", "language", "grade", "country", "region", "signin_country", "cohort"])
  public_reading = merge(local.names, { fold = file("${local.analytics}/views/public/fold.sql") })
}

resource "google_project_service" "analytics" {
  for_each = toset([
    "logging.googleapis.com",
    "bigquery.googleapis.com",
    # The service that runs the nightly query.
    "bigquerydatatransfer.googleapis.com",
  ])

  service = each.value

  # Switching an API off is not the opposite of switching it on: it would break
  # whatever else in the project still uses it.
  disable_on_destroy = false
}

# The raw lines, 62 days of them: enough to count the month before last again
# on its last night. Analytics, once on, cannot be switched off.
resource "google_logging_project_bucket_config" "activity" {
  project          = var.project_id
  location         = var.region
  bucket_id        = "activity"
  retention_days   = 62
  enable_analytics = true
  description      = "The lines the children are counted from, kept 62 days and then counted for good."

  depends_on = [google_project_service.analytics]
}

# Only the lines the children are counted from are copied. They stay in the
# default bucket as well, where the report reads every line of the service.
resource "google_logging_project_sink" "activity" {
  name        = "activity"
  destination = "logging.googleapis.com/${google_logging_project_bucket_config.activity.id}"
  description = "The lines the children are counted from."
  filter = join(" AND ", [
    "resource.type=\"cloud_run_revision\"",
    "resource.labels.service_name=\"${var.service_name}\"",
    "log_id(\"run.googleapis.com/stdout\")",
    "jsonPayload.message=(${join(" OR ", [for event in local.events : "\"${event}\""])})",
  ])
}

# The bucket as BigQuery reads it: a dataset that points at the lines and holds
# no copy of them.
resource "google_logging_linked_dataset" "activity" {
  link_id     = "activity_logs"
  bucket      = google_logging_project_bucket_config.activity.id
  description = "The lines the children are counted from, as BigQuery reads them."
}

# A query reads only datasets that are kept in one place. A bucket in a region
# is read in that region, which is where the counts are kept; the platform
# does not promise it, so every plan looks.
check "linked_dataset_location" {
  data "google_bigquery_dataset" "logs" {
    project    = var.project_id
    dataset_id = google_logging_linked_dataset.activity.link_id
  }

  assert {
    condition     = lower(data.google_bigquery_dataset.logs.location) == lower(var.region)
    error_message = "The linked dataset of the activity bucket is in ${data.google_bigquery_dataset.logs.location}, not in ${var.region}: the nightly query cannot read it beside the counts."
  }
}

resource "google_bigquery_dataset" "impact" {
  dataset_id  = "impact"
  location    = var.region
  description = "The counts of the children, kept for years. No table holds a name any child is counted under, or anything else that could name one."

  depends_on = [google_project_service.analytics]
}

resource "google_bigquery_dataset" "private" {
  dataset_id  = "impact_private"
  location    = var.region
  description = "The counts as exact numbers, for the owner's report alone."

  depends_on = [google_project_service.analytics]
}

resource "google_bigquery_dataset" "public" {
  dataset_id  = "impact_public"
  location    = var.region
  description = "The counts as anyone may see them: closed months, no group of fewer than ten children, every number to the nearest five."

  depends_on = [google_project_service.analytics]
}

# The tables kept for years. Each schema is written out, so a query that tried
# to keep a column of its own — a child's name among them — would fail rather
# than keep it, and a change of schema that would replace a table is refused.
resource "google_bigquery_table" "counts" {
  for_each = local.tables

  dataset_id          = google_bigquery_dataset.impact.dataset_id
  table_id            = each.key
  schema              = file("${local.analytics}/tables/${each.key}.json")
  deletion_protection = true
}

# Who the nightly query runs as: it reads the lines, writes the counts and runs
# its own jobs, and nothing else.
resource "google_service_account" "counter" {
  account_id   = "${var.service_name}-impact"
  display_name = "Counts of how the ${var.service_name} service is used"

  depends_on = [google_project_service.analytics]
}

resource "google_project_iam_member" "counter_jobs" {
  project = var.project_id
  role    = "roles/bigquery.jobUser"
  member  = "serviceAccount:${google_service_account.counter.email}"
}

resource "google_bigquery_dataset_iam_member" "counter_reads_lines" {
  dataset_id = google_logging_linked_dataset.activity.link_id
  role       = "roles/bigquery.dataViewer"
  member     = "serviceAccount:${google_service_account.counter.email}"
}

resource "google_bigquery_dataset_iam_member" "counter_writes_counts" {
  dataset_id = google_bigquery_dataset.impact.dataset_id
  role       = "roles/bigquery.dataEditor"
  member     = "serviceAccount:${google_service_account.counter.email}"
}

# Every night, half past one in UTC — not on the hour, which the platform may
# run twice — the counts are made again from the lines. The script needs no
# date of its own: it counts every period the lines still hold whole, so a
# night that did not run is made up for by the next. The rights it runs with
# take a while to spread, and a first apply may have to be run again.
resource "google_bigquery_data_transfer_config" "nightly" {
  display_name         = "${var.service_name} counts"
  location             = var.region
  data_source_id       = "scheduled_query"
  schedule             = "every day 01:30"
  service_account_name = google_service_account.counter.email

  params = {
    query = templatefile("${local.analytics}/nightly.sql", local.names)
  }

  depends_on = [
    google_project_service.analytics,
    google_project_iam_member.counter_jobs,
    google_bigquery_dataset_iam_member.counter_reads_lines,
    google_bigquery_dataset_iam_member.counter_writes_counts,
    google_bigquery_table.counts,
  ]
}

# The views. A view keeps no data, so it may be replaced; and what a view reads
# is named inside its query, where Terraform does not see it, so the order is
# written out.
resource "google_bigquery_table" "private" {
  for_each = local.private_views

  dataset_id          = google_bigquery_dataset.private.dataset_id
  table_id            = each.key
  deletion_protection = false

  view {
    query          = templatefile("${local.analytics}/views/private/${each.key}.sql", local.names)
    use_legacy_sql = false
  }

  depends_on = [google_bigquery_table.counts]
}

resource "google_bigquery_table" "closed_months" {
  dataset_id          = google_bigquery_dataset.public.dataset_id
  table_id            = "closed_months"
  deletion_protection = false

  view {
    query          = templatefile("${local.analytics}/views/public/closed_months.sql", local.names)
    use_legacy_sql = false
  }

  depends_on = [google_bigquery_table.counts]
}

# The children of each month told apart by one thing, a view to each.
resource "google_bigquery_table" "learners_by" {
  for_each = local.dimensions

  dataset_id          = google_bigquery_dataset.public.dataset_id
  table_id            = "learners_by_${each.key}"
  deletion_protection = false

  view {
    query          = templatefile("${local.analytics}/views/public/learners_by.sql", merge(local.public_reading, { dimension = each.key }))
    use_legacy_sql = false
  }

  depends_on = [google_bigquery_table.closed_months]
}

resource "google_bigquery_table" "public" {
  for_each = local.public_views

  dataset_id          = google_bigquery_dataset.public.dataset_id
  table_id            = each.key
  deletion_protection = false

  view {
    query          = templatefile("${local.analytics}/views/public/${each.key}.sql", local.public_reading)
    use_legacy_sql = false
  }

  depends_on = [google_bigquery_table.closed_months, google_bigquery_table.learners_by]
}
