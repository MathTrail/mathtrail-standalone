# The APIs this project is allowed to use. Service Usage and Resource Manager
# are not in the list because they are what enables the rest: they have to be on
# before this configuration can be applied at all.

locals {
  apis = [
    # The service itself, and the custom domain in front of it.
    "run.googleapis.com",
    # Where its images are kept.
    "artifactregistry.googleapis.com",
    # Where the sealing key and the client secret are kept.
    "secretmanager.googleapis.com",
    # The child's profile is a file in the parent's own Drive.
    "drive.googleapis.com",
    # The identity a deployment arrives as: a federated pool, the token exchange
    # that turns a GitHub token into a Google one, and the impersonation that
    # follows.
    "iam.googleapis.com",
    "sts.googleapis.com",
    "iamcredentials.googleapis.com",
    # The spend alert.
    "billingbudgets.googleapis.com",
  ]
}

resource "google_project_service" "enabled" {
  for_each = toset(local.apis)

  service = each.value

  # Switching an API off is not the opposite of switching it on: it would break
  # whatever else in the project still uses it, and the resources that needed it
  # here are gone by the time this would run.
  disable_on_destroy = false
}
