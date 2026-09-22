# Running your own copy

MathTrail is MIT-licensed and holds nothing of its own: every deployment is one Cloud Run service, one Artifact Registry repository, two secrets and a domain, in a Google Cloud project you control. This page is how the infrastructure is created — what is described in Terraform, what has to be done by hand, and in which order.

The name is not part of the licence: a copy that is not this project is named differently.

## What you need

- A Google Cloud project with billing attached, and a billing account you may create budgets on.
- A domain you have verified ownership of, and the ability to add a DNS record to it. The service answers on a host of its own — `mcp.example.com` — while the site lives wherever you publish it.
- A GitHub repository, if you want deployments to run from it.
- The development container of this repository: it carries Terraform, gcloud and everything else at the exact versions this configuration is checked against. On arm64 machines gcloud is not in the image, because its archive for that architecture ships no Python interpreter; run the same commands from Google's `google-cloud-cli` container image instead.

## 1. Sign in

Everything below runs inside the development container. The sign-in survives rebuilds: the credentials live in a volume of their own.

```bash
gcloud auth login
gcloud auth application-default login   # the credentials Terraform reads
gcloud config set project PROJECT_ID
```

## 2. Bootstrap what Terraform cannot create

Two APIs have to be on before a configuration that enables APIs can be applied at all, and the bucket that holds the state cannot be described by the configuration whose state it holds.

```bash
gcloud services enable serviceusage.googleapis.com cloudresourcemanager.googleapis.com

gcloud storage buckets create gs://BUCKET \
    --location=us-central1 \
    --uniform-bucket-level-access \
    --public-access-prevention
gcloud storage buckets update gs://BUCKET --versioning
```

Versioning is what lets a state file that was written badly be rolled back to the one before it.

## 3. Point the domain at Cloud Run

Do this before the first apply. A domain mapping is created only for a domain whose ownership is already verified, and its certificate is issued only once the DNS record resolves — so the record goes in first:

```
mcp.example.com.   CNAME   ghs.googlehosted.com.
```

Ownership is verified once, for the whole domain, in Google Search Console, with the account you are signed in as.

## 4. Configure

```bash
cd infra/terraform
cp terraform.tfvars.example terraform.tfvars   # fill it in
cp backend.hcl.example backend.hcl             # the bucket from step 2
terraform init -backend-config=backend.hcl
```

Neither file is committed: `terraform.tfvars` names one particular deployment, and `backend.hcl` names where its state lives.

## 5. Apply, in two passes

The secrets are created empty — their values never pass through Terraform, and never appear in the state. A revision cannot start before they hold something, so the first apply stops at the secrets:

```bash
terraform apply \
    -target=google_secret_manager_secret.seal_key \
    -target=google_secret_manager_secret.google_client_secret

# 32 random bytes, base64-encoded: the key the tokens and the task answers are sealed with
head -c 32 /dev/urandom | base64 | tr -d '\n' \
    | gcloud secrets versions add mathtrail-seal-key --data-file=-

# the client secret of the Google OAuth client this deployment signs in with
printf '%s' 'CLIENT_SECRET' | gcloud secrets versions add mathtrail-google-client-secret --data-file=-

terraform apply
```

The full apply creates everything else, including the domain mapping, and waits for the certificate — fifteen minutes is normal, and the DNS record from step 3 is what it is waiting on. If it gives up before the certificate is issued, run `terraform apply` again.

`terraform output` then prints the platform's own address for the service, where images are pushed, the identity a deployment borrows and the provider it presents its token to.

## 6. Leave one entrance

Once the domain answers, withdraw the platform's address, so that there is one way in and one issuer of tokens:

```hcl
disable_default_url = true
```

This is a step of its own and not a default, because Cloud Run asks for the domain to be mapped before its own address is withdrawn — and until the certificate exists, that address is the only way to reach anything at all.

A deployment that has no domain of its own leaves both settings alone, sets `create_domain_mapping = false`, and puts the platform's address in `public_host` — which is known only after the first apply, so it takes two.

## 7. The first image

Until an image of this service is deployed, the service answers with Google's placeholder container: the configuration defaults to it so that a fresh project applies cleanly. From then on the deployment owns which image is served and Terraform ignores that one field, so an apply never puts the placeholder back and the `image` variable matters only when a service is created from nothing.

## 8. Deploying

A deployment builds the image of one commit, pushes it to the repository created in step 5, rolls a revision of the service by digest, and then asks the service which commit it is serving. It starts on a push to `main`, or by hand from any branch. It uses no key and no secret: the workflow mints a federated token for itself, and everything it needs to know is six repository variables, each a copy of one output of the configuration that created the deployment target.

| Repository variable | Terraform output |
|---|---|
| `GCP_WORKLOAD_IDENTITY_PROVIDER` | `workload_identity_provider` |
| `GCP_DEPLOY_SERVICE_ACCOUNT` | `deploy_service_account` |
| `GCP_IMAGE_REPOSITORY` | `image_repository` |
| `GCP_SERVICE` | `service_name` |
| `GCP_REGION` | `region` |
| `PUBLIC_URL` | `public_url` |

Read them with `terraform output` and put them in the repository's settings, under Actions variables — not secrets, because none of them is one. Until the custom domain answers, `PUBLIC_URL` is whatever address the service is actually reachable at, or the last step of every deployment fails.

The checks are not run again here: they gate the pull request that a merge comes from, so protect `main` and let them do it there. A commit pushed straight to `main` is deployed without them.

The same three steps run from a laptop, which is what an emergency deployment or a first bring-up looks like:

```bash
image=$(just ci-image-push us-central1-docker.pkg.dev/PROJECT_ID/mathtrail/server)
just ci-deploy mathtrail us-central1 "$image"
just ci-smoke https://mcp.example.com
```

Rolling back is a revision, never an apply — Terraform does not know which digest is live, and Cloud Run keeps every revision that ever served:

```bash
gcloud run revisions list --service=mathtrail --region=us-central1
gcloud run services update-traffic mathtrail --region=us-central1 --to-revisions=REVISION=100
```

Running the deployment again at an older commit does the same thing the long way round, and leaves the registry with an image for that commit.

## 9. The Google sign-in

The parent signs in with Google, and the service asks Google for exactly two things: a verified identifier, and room for one file of its own in the parent's Drive. None of it is in Terraform — it is configured by hand in the Google Cloud console, under **Google Auth Platform**, on its four pages.

**Branding.** The name people see on the consent screen, a support email, and three links: the home page, the privacy policy and the terms — for this deployment `https://mathtrail.app/en/`, `https://mathtrail.app/en/privacy/` and `https://mathtrail.app/en/terms/`. Add the top private domain of those links, `mathtrail.app`, as an authorized domain: one entry covers both the site on the apex and the service on its subdomain. Its ownership is already verified from step 3.

**Audience.** User type **External**, publishing status **In production**. Not Testing: there a consent expires seven days after it is given and takes the refresh token with it, so every parent would be signed out once a week.

**Data access.** Two scopes and no others: `openid`, which is what makes the identifier verified, and `https://www.googleapis.com/auth/drive.file`, which reaches only the files the app itself created plus anything the parent hands it explicitly. Both are non-sensitive, so publishing needs no app verification at all; brand verification is the separate, lighter process that makes the app's own name and logo appear on the consent screen instead of the project's name.

**Clients.** One client, type **Web application**, with a single authorized redirect URI — `https://<your host>/oauth/callback`, no trailing slash. No authorized JavaScript origins: the sign-in is a redirect the service performs, never a script inside a page. The client id goes into `terraform.tfvars`; the client secret goes into Secret Manager in step 5.

Neither the project nor this client is ever recreated. A parent grants `drive.file` to that client in that project, and the grant cannot be moved: a new client sees none of the files parents have already given it, and every child's profile becomes unreachable through the app.

## The first deployment, as a checklist

Once, in this order. Each line points at the section that has the commands.

- [ ] A Google Cloud project with billing attached, and the right to create a budget on that billing account.
- [ ] [Sign in](#1-sign-in) from inside the container.
- [ ] [Bootstrap](#2-bootstrap-what-terraform-cannot-create) the two APIs and the state bucket.
- [ ] [The domain](#3-point-the-domain-at-cloud-run): ownership verified, the CNAME created. Early, because the certificate is issued only once that record resolves, and that can take a day.
- [ ] [The Google sign-in](#9-the-google-sign-in): the consent screen in production and one Web client. Keep the client id and the client secret.
- [ ] [Configure](#4-configure) Terraform: `terraform.tfvars` with that client id, `backend.hcl`, `terraform init`.
- [ ] [Apply in two passes](#5-apply-in-two-passes): the secrets, their values by hand, then everything else.
- [ ] [The deployment variables](#8-deploying): six repository variables, read with `terraform output`.
- [ ] The first deployment: a merge to `main`, or the workflow started by hand from a branch.
- [ ] [Leave one entrance](#6-leave-one-entrance), once the domain answers: `disable_default_url = true`, and apply again.

It is not finished until all four of these say so:

```bash
# the domain answers, and answers as the commit that was deployed
just ci-smoke https://mcp.example.com

# the same version and commit in the line the service logs as it starts
gcloud run services logs read mathtrail --region=us-central1 --limit=20

# the spend alert exists and names this project
gcloud billing budgets list --billing-account=01ABCD-234567-89EFGH

# both cleanup policies are in force, and cleanupPolicyDryRun is false
gcloud artifacts repositories describe mathtrail --location=us-central1
```

## The variables

| Variable | Default | What it is |
|---|---|---|
| `project_id` | — | The project everything lives in |
| `billing_account` | — | The account the spend alert is created on |
| `public_host` | — | The host the service answers on and calls itself |
| `google_oauth_client_id` | — | The Google OAuth client of the sign-in |
| `github_repository` | — | `owner/name` of the repository allowed to deploy |
| `github_owner_id` | — | The numeric id of that owner |
| `region` | `us-central1` | Has to offer domain mappings and be a first-tier region |
| `service_name` | `mathtrail` | Names the service, its images, its secrets and its identities |
| `image` | a placeholder | The image to serve, always by digest |
| `max_instances` | `3` | The ceiling on instances running at once |
| `concurrency` | `80` | Requests one instance serves at a time |
| `cpu`, `memory` | `1`, `512Mi` | Per instance, allocated only while serving |
| `request_timeout` | `60s` | Before the platform cuts a request off |
| `disable_default_url` | `false` | Whether the platform's own address stops resolving; turned on once the domain answers |
| `create_domain_mapping` | `true` | Whether `public_host` is mapped onto the service |
| `seal_key_version` | `1` | The secret version the service seals with |
| `seal_key_previous_version` | empty | The version still accepted while a key is being rotated |
| `google_client_secret_version` | `1` | The secret version the sign-in authenticates with |
| `settings` | `{}` | Extra environment variables — ceilings and timeouts, never a secret |
| `budget_amount`, `budget_currency` | `1`, `USD` | Where the spend alert fires |
| `keep_images` | `5` | Image versions kept whatever their age |
| `image_max_age` | `30d` | When an older version is deleted |

## What Terraform does not own

- **The project and its billing.** Both exist before the first apply.
- **The secret values.** Added by hand, so that no state file ever holds one.
- **The Google OAuth client and the consent screen.** Created in the Google console; the client id goes into `terraform.tfvars` and the client secret into Secret Manager.
- **Domain ownership and DNS.** Verified and configured with the registrar.
- **The image.** Built and deployed by the workflow; the variable here is only what a service created from nothing starts with.

## What can cost money

The intent is $0, and inside the free allowance it is $0 — but the allowance is what makes it free, not the configuration. What to watch:

- **Cloud Run** beyond the monthly free requests, vCPU-seconds and GiB-seconds, or beyond the free egress from North America. `min_instance_count` is 0 and CPU is allocated only during requests, so an idle service costs nothing; a service that is being used a great deal does not stay free.
- **Artifact Registry** beyond half a gigabyte of images. The cleanup policies keep the repository small; a deployment that pushes many images a day should check that they are working.
- **Secret Manager** beyond six active versions or the monthly free accesses. A version is read once per instance start, and the rotation keeps at most three versions live.
- **Cloud Logging** beyond the free monthly ingestion.
- **A region that is not first-tier**, where the free allowance does not apply.
- **A load balancer**, if the domain mapping is ever replaced by one. Domain mappings cost nothing; a forwarding rule is billed by the hour whether anybody visits or not.

The spend alert warns, it does not stop anything: Google has no switch that halts a project at a number. Treat the first alert as a fault to investigate.

**A caveat about the domain mapping.** Google labels Cloud Run domain mappings a preview feature and says they are not production-ready, citing latency. They are the only way to put a custom domain in front of Cloud Run without paying for a load balancer, which is why they are used here, and the alternative — a global external Application Load Balancer — is a monthly bill rather than a free tier.

## Rotating the sealing key

Both keys are versions of the same secret. Add the new version, make it the current one, keep the version that was current as the previous one, apply, and only then disable what fell off the end:

```bash
head -c 32 /dev/urandom | base64 | tr -d '\n' \
    | gcloud secrets versions add mathtrail-seal-key --data-file=-
```

```hcl
seal_key_version          = "3"
seal_key_previous_version = "2"
```

Tokens sealed with the previous key keep working until they expire on their own, so nobody is signed out by a rotation.

## Removing a deployment

`terraform destroy` removes everything this configuration created, including the secrets and every version in them: copy the sealing key out first if anything sealed with it still matters. The APIs stay enabled, because switching one off breaks whatever else in the project still uses it. The project, the state bucket, the DNS record and the OAuth client were never described here and stay as they are.
