# Everything a deployment has to say about itself. A variable with no default is
# one only the person deploying can answer; every other one carries the value
# this service is meant to run with.

variable "project_id" {
  description = "The Google Cloud project that holds every resource here."
  type        = string
}

variable "billing_account" {
  description = "The billing account the project pays through, as an id of the form 01ABCD-234567-89EFGH. The spend alert is created on it."
  type        = string
}

variable "region" {
  description = "Where the service, its images and its domain mapping live. It has to be a region that offers domain mappings, and a first-tier region for the free allowance to apply."
  type        = string
  default     = "us-central1"
}

variable "public_host" {
  description = "The host the service is reached at and calls itself, with no scheme and no path: the issuer of its tokens and the address the sign-in returns to."
  type        = string

  validation {
    condition     = can(regex("^[a-z0-9]([a-z0-9-]*[a-z0-9])?(\\.[a-z0-9]([a-z0-9-]*[a-z0-9])?)+$", var.public_host))
    error_message = "public_host is a bare host name, such as mcp.example.com."
  }
}

variable "google_oauth_client_id" {
  description = "The Google OAuth client the sign-in uses. Not a secret: it travels in every authorization request."
  type        = string
}

variable "github_repository" {
  description = "The repository whose workflows may deploy, as owner/name. Only tokens issued to it are accepted."
  type        = string

  validation {
    condition     = can(regex("^[^/]+/[^/]+$", var.github_repository))
    error_message = "github_repository is written as owner/name."
  }
}

variable "github_owner_id" {
  description = "The numeric id of the account or organisation that owns the repository, from https://api.github.com/users/<owner>. A name can be given up and taken by somebody else; a number cannot."
  type        = string

  validation {
    condition     = can(regex("^[0-9]+$", var.github_owner_id))
    error_message = "github_owner_id is the numeric id, not the name."
  }
}

variable "service_name" {
  description = "The name of the Cloud Run service, and the prefix of everything named after it."
  type        = string
  default     = "mathtrail"
}

variable "image" {
  description = "The container image to serve, by digest. The default is a placeholder that answers until the first image of this service is deployed."
  type        = string
  default     = "us-docker.pkg.dev/cloudrun/container/hello@sha256:be0a21e5d7036cc60741cf60ea3b68e169cc2d00f3ad256c94536f32b0a0755c"

  validation {
    condition     = can(regex("@sha256:[0-9a-f]{64}$", var.image))
    error_message = "The image is named by digest, never by tag: a tag can be moved under a running service."
  }
}

variable "max_instances" {
  description = "How many instances may run at once. Each one keeps its own request-rate counter, so the effective rate ceiling is this many times the configured one."
  type        = number
  default     = 3
}

variable "concurrency" {
  description = "How many requests one instance serves at a time."
  type        = number
  default     = 80
}

variable "cpu" {
  description = "CPU per instance. Allocated only while a request is being served."
  type        = string
  default     = "1"
}

variable "memory" {
  description = "Memory per instance. The binary carries its content and its solver sandbox, and holds nothing else between requests."
  type        = string
  default     = "512Mi"
}

variable "request_timeout" {
  description = "How long a single request may take before the platform cuts it off. Nothing the service does takes anywhere near this long; the model's own thinking happens in the chat, not here."
  type        = string
  default     = "60s"
}

variable "disable_default_url" {
  description = "Whether the platform's own address for the service stops resolving, leaving the custom domain as the only entrance. It is turned on after that domain is mapped and answering: the platform asks for the mapping to exist first, and until its certificate is issued there is no other way in."
  type        = bool
  default     = false
}

variable "create_domain_mapping" {
  description = "Whether to map public_host onto the service. A deployment that has no domain of its own leaves it off and is reached at the platform's address."
  type        = bool
  default     = true
}

variable "seal_key_version" {
  description = "The version of the sealing-key secret the service seals with."
  type        = string
  default     = "1"
}

variable "seal_key_previous_version" {
  description = "The version of the sealing-key secret that is still accepted for unsealing during a rotation. Empty the rest of the time, and then the service is not given one at all."
  type        = string
  default     = ""
}

variable "google_client_secret_version" {
  description = "The version of the Google client-secret secret the service authenticates with."
  type        = string
  default     = "1"
}

variable "settings" {
  description = "Extra environment variables, for the ceilings and timeouts the binary otherwise defaults to. Never a secret: these values are readable in the state and in the deployed revision."
  type        = map(string)
  default     = {}
}

variable "budget_amount" {
  description = "The monthly spend, in whole units of budget_currency, that the alert is measured against. This is an alert and not a cap: nothing stops a project at a number."
  type        = string
  default     = "1"
}

variable "budget_currency" {
  description = "The currency of the spend alert. It has to be the currency of the billing account."
  type        = string
  default     = "USD"
}

variable "keep_images" {
  description = "How many of the most recent image versions survive the cleanup, whatever their age."
  type        = number
  default     = 5
}

variable "image_max_age" {
  description = "How old an image version may get before the cleanup deletes it, unless it is one of the most recent ones."
  type        = string
  default     = "30d"
}
