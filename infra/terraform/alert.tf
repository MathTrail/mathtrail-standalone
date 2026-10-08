# The load alert. The spend alert hears of a burst only once Billing has
# counted it, hours later; this one reads the platform's own count of the
# service's requests, which costs nothing to keep, and mails the operator while
# the burst lasts and again once it is over. A deployment that names no
# operator gets neither the channel nor the alert.

resource "google_monitoring_notification_channel" "operator" {
  count = var.operator_email == "" ? 0 : 1

  display_name = "${var.service_name} operator"
  type         = "email"
  labels = {
    email_address = var.operator_email
  }

  depends_on = [google_project_service.enabled]
}

resource "google_monitoring_alert_policy" "load" {
  count = var.operator_email == "" ? 0 : 1

  display_name = "${var.service_name} load"
  combiner     = "OR"
  severity     = "WARNING"

  conditions {
    display_name = "Requests a minute, on average over five minutes"

    condition_threshold {
      filter = join(" AND ", [
        "resource.type=\"cloud_run_revision\"",
        "resource.labels.service_name=\"${var.service_name}\"",
        "metric.type=\"run.googleapis.com/request_count\"",
      ])

      # The rate over the last five minutes, every revision and every answer
      # summed: an average that rolls, so a burst of seconds mails nobody.
      aggregations {
        alignment_period     = "300s"
        per_series_aligner   = "ALIGN_RATE"
        cross_series_reducer = "REDUCE_SUM"
      }

      comparison      = "COMPARISON_GT"
      threshold_value = var.load_alert_per_minute / 60

      # A minute over the line before it fires: Cloud Monitoring takes a rule
      # for missing points only in a condition that waits a minute at least.
      # No requests write no points, and that is no load, so the incident
      # closes rather than waiting for data that will not come.
      duration                = "60s"
      evaluation_missing_data = "EVALUATION_MISSING_DATA_INACTIVE"
    }
  }

  alert_strategy {
    auto_close           = "1800s"
    notification_prompts = ["OPENED", "CLOSED"]
  }

  notification_channels = [google_monitoring_notification_channel.operator[0].name]

  documentation {
    subject   = "${var.service_name}: more than ${var.load_alert_per_minute} requests a minute"
    mime_type = "text/markdown"
    content   = <<-EOT
      The service has had more than ${var.load_alert_per_minute} requests a minute on average over the last five minutes. A card waiting for a task asks fifteen times a minute, so that is some ${floor(var.load_alert_per_minute / 15)} families at once, or somebody who is not a family.

      - `just usage 1h`: what Cloud Run counted of the last hour against the free tier, and the most instances that served in one minute.
      - `just report 1h`: the busiest minute, the limits the service reached and how its tools answered.

      Every instance still holds its paces, and at most ${var.max_instances} run at once. A second mail comes once the load is back under the line.
    EOT
  }

  depends_on = [google_project_service.enabled]
}
