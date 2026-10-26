# What the platform's own record of requests keeps. Cloud Run writes the whole
# address of every request it serves into the project's log, and the address
# the sign-in is asked at carries what a chat and Google put in its query:
# Google's one-time code, the domain of a Workspace account, and the email
# address a chat suggests signing in with. The service's own line of each
# request already keeps only the parameters defined to hold no data, so the
# platform's record of a request whose address carries a query is not kept at
# all. Of the service's own addresses only the sign-in's carry one; a request
# without one, a path a host asked for in vain among them, is kept as before.

resource "google_logging_project_exclusion" "requests_with_a_query" {
  name        = "${var.service_name}-requests-with-a-query"
  description = "Cloud Run's record of a request whose address carries a query, as the sign-in's do: a one-time code, a Workspace domain, an email address."
  filter = join(" AND ", [
    "resource.type=\"cloud_run_revision\"",
    "resource.labels.service_name=\"${var.service_name}\"",
    "log_id(\"run.googleapis.com/requests\")",
    "httpRequest.requestUrl:\"?\"",
  ])

  depends_on = [google_project_service.enabled]
}
