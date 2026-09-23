# Running your own copy

MathTrail is MIT-licensed and holds nothing of its own: every deployment is one Cloud Run service, one Artifact Registry repository, two secrets and a domain, in a Google Cloud project you control. This page is how that gets created — and almost none of it is done by hand. The repository describes the deployment, and a workflow delivers it.

The name is not part of the licence: a copy that is not this project is named differently.

## What is done by hand, ever

Five things, once, because no API this repository uses can do them:

1. **The bootstrap**, below: the project, the bucket its state lives in and the identity the pipeline signs in as. A pipeline cannot create the door it walks in by. Once — and again on the day that identity needs a right it was not given, which is its own section further down.
2. **The Google consent screen and one OAuth client.** Google has no API for either.
3. **Two DNS records**, at whatever registrar holds the domain, and one click that makes the deployment identity a verified owner of it.
4. **The OAuth client's secret**, pasted into the repository's secrets. It is the one value that exists nowhere but the Google console.
5. **Storage for the traces**, once the service is deployed: in **Trace Explorer**, a banner says trace storage is not initialized, and its **Enable** button creates the observability bucket named `_Trace`. Nothing else creates it — spans from Cloud Run deliberately do not, `gcloud` cannot, and the Terraform provider has no resource for it; the only other way is the Observability API directly (`POST https://observability.googleapis.com/v1/projects/PROJECT_ID/locations/LOCATION/buckets?bucketId=_Trace`, which needs `roles/observability.editor`). Until it exists the endpoint accepts every span and stores none, and nothing anywhere says so. Enabling it back-fills the last hour, so the spans sent while it was missing are not lost.

Everything else — enabling APIs, the registry and its cleanup, the secrets, the service, the domain mapping, the spend alert, the image and every roll-out after it — happens when a change reaches the main branch, or when the workflow is started by hand from a branch.

## What you need

- A Google Cloud project with billing attached, or the right to create one, and a billing account you may create budgets on.
- A domain, and the ability to add a record to it. The service answers on a host of its own — `mcp.example.com` — while the site lives wherever you publish it.
- The development container of this repository, or a browser with Cloud Shell: the bootstrap is the only step that needs a shell at all.

## 1. Say what the deployment is

Three files in the repository, and nothing about the deployment lives anywhere else:

| File | What it holds |
|---|---|
| `infra/terraform/prod.auto.tfvars` | project, region, host, OAuth client id, repository and its owner id, pool id — everything except the billing account |
| `infra/terraform/backend.hcl` | the bucket the state lives in |
| `infra/ci.env` | the two facts a workflow needs before it can sign in — filled from what the bootstrap prints |

None of it is secret: the client id travels in every authorization request, and the federated pool is guarded by a condition on the repository rather than by the secrecy of its name. Changing where this repository deploys is a pull request, which is the point.

## 2. Bootstrap, once

```bash
TF_VAR_billing_account=01ABCD-234567-89EFGH just bootstrap
```

Or, with no local tooling at all, paste the contents of `infra/bootstrap.sh` into Cloud Shell. It is safe to run again: every step checks before it creates.

It creates the project and links billing, enables the handful of APIs without which nothing else can be created, makes the state bucket with versioning, creates the federated pool and its GitHub provider, and creates the identity the infrastructure is applied as — with an enumerated set of roles rather than Owner. It prints two lines for `infra/ci.env`. Commit them.

## 3. The Google sign-in

The parent signs in with Google, and the service asks Google for exactly two things: a verified identifier, and room for one file of its own in the parent's Drive. It is configured in the Google Cloud console, under **Google Auth Platform**, on its four pages.

**Branding.** The name people see on the consent screen, a support email, and three links: the home page, the privacy policy and the terms — for this deployment `https://mathtrail.app/en/`, `https://mathtrail.app/en/privacy/` and `https://mathtrail.app/en/terms/`. Add the top private domain of those links, `mathtrail.app`, as an authorized domain: one entry covers both the site on the apex and the service on its subdomain.

**Audience.** User type **External**, publishing status **In production**. Not Testing: there a consent expires seven days after it is given and takes the refresh token with it, so every parent would be signed out once a week.

**Data access.** Two scopes and no others: `openid`, which is what makes the identifier verified, and `https://www.googleapis.com/auth/drive.file`, which reaches only the files the app itself created plus anything the parent hands it explicitly. Both are non-sensitive, so publishing needs no app verification at all; brand verification is the separate, lighter process that makes the app's own name and logo appear on the consent screen instead of the project's name.

**Clients.** One client, type **Web application**, with a single authorized redirect URI — `https://<your host>/oauth/callback`, no trailing slash. No authorized JavaScript origins: the sign-in is a redirect the service performs, never a script inside a page. The client id goes into `prod.auto.tfvars`.

Neither the project nor this client is ever recreated. A parent grants `drive.file` to that client in that project, and the grant cannot be moved: a new client sees none of the files parents have already given it, and every child's profile becomes unreachable through the app.

## 4. The domain

1. Verify ownership of the domain in [Search Console](https://search.google.com/search-console), with the TXT record it asks for.
2. At the registrar: the service's host as a **CNAME** to `ghs.googlehosted.com.`
3. In Search Console, under the domain's settings, add `mathtrail-tf@<project>.iam.gserviceaccount.com` — the identity the configuration is applied as — as a verified owner. Cloud Run maps a domain only for an owner of it, and that identity is the one creating the mapping; without this the delivery gets as far as the domain and stops.

Do all three before the first delivery: the certificate is issued only once the record resolves, and that can take a day.

## 5. The two repository secrets

Settings → Secrets and variables → Actions → Secrets:

| Secret | What it is |
|---|---|
| `GOOGLE_OAUTH_CLIENT_SECRET` | The OAuth client's secret. Used exactly once, on the first apply of a deployment, to put the first version into Secret Manager; every apply after that leaves it alone. |
| `TF_VAR_billing_account` | The id of the billing account, of the form `01ABCD-234567-89EFGH`. It is the one fact about the deployment that is not written down beside the others, because it names the account that pays and this repository is public. Terraform reads variables from `TF_VAR_`-prefixed environment variables by itself. |

The bootstrap needs the second one too, so export it there: `TF_VAR_billing_account=01ABCD-234567-89EFGH just bootstrap`.

The key everything is sealed with is in neither list: the pipeline generates 32 random bytes on the first apply and writes them straight into Secret Manager. No person, no state file and no log ever sees that value.

## 6. Deliver

A push to `main` — or the workflow started by hand from any branch — runs three jobs:

1. **The cloud, as described.** `terraform apply`, every time, so the deployment always matches the branch it was cut from. On the very first run this is what creates everything; afterwards it is usually a no-op that takes half a minute.
2. **The image of this commit.** Built, pushed to the registry with the commit as its tag, and rolled out **by digest** — a tag can be moved afterwards and a digest cannot.
3. **The check.** `/health` is asked which commit it is serving, and the delivery fails unless the answer is the commit that was just built.

A pull request that touches `infra/` gets a `terraform plan` in its summary instead, so what the cloud is about to become is reviewable before the merge.

## 7. Leave one entrance

Once the domain answers, withdraw the platform's own address for the service, so that there is one way in and one issuer of tokens: set `disable_default_url = true` in `prod.auto.tfvars` and merge. It is a step of its own because Cloud Run asks for the domain to be mapped before its own address is withdrawn — and until the certificate exists, that address is the only way to reach anything at all.

A deployment with no domain of its own instead sets `create_domain_mapping = false` and puts the platform's address in `public_host`, which is known only after the first delivery — so it takes two.

## The first delivery, as a checklist

- [ ] `prod.auto.tfvars` and `backend.hcl` filled in, and a project id nobody has taken.
- [ ] `just bootstrap`, and the two lines it prints committed to `infra/ci.env`.
- [ ] The consent screen published and a Web client created; its client id in `prod.auto.tfvars`.
- [ ] The domain verified, its CNAME created, and the deployment identity added as a verified owner.
- [ ] Trace storage enabled from the banner in Trace Explorer, after the first delivery.
- [ ] Both repository secrets: `GOOGLE_OAUTH_CLIENT_SECRET` and `TF_VAR_billing_account`.
- [ ] Merged to `main`, and all three jobs green.
- [ ] `disable_default_url = true` merged, once the domain answers.

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
| `billing_account` | — | The account the spend alert is created on; comes from `TF_VAR_billing_account`, not from a file |
| `public_host` | — | The host the service answers on and calls itself |
| `google_oauth_client_id` | — | The Google OAuth client of the sign-in |
| `github_repository` | — | `owner/name` of the repository allowed to deploy |
| `github_owner_id` | — | The numeric id of that owner; the pool's condition is built from it |
| `workload_identity_pool_id` | `mathtrail-github` | The pool created by the bootstrap, in which a workflow's token is exchanged |
| `region` | `us-central1` | Has to offer domain mappings and be a first-tier region |
| `service_name` | `mathtrail` | Names the service, its images, its secrets and its identities |
| `image` | a placeholder | What a service created from nothing starts with; after that the delivery owns the field |
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

## What nothing in this repository owns

- **The project's billing.** Linked by the bootstrap, owned by whoever pays.
- **The values of the secrets.** One is generated by the pipeline and read by nobody; the other is pasted once. Neither is ever in the state.
- **The Google OAuth client and the consent screen.** No API exists; they are the reason step 3 is done by hand.
- **Domain ownership and DNS.** At the registrar and in Search Console.
- **The pool the pipeline signs in through.** Created by the bootstrap, deliberately outside Terraform.

## What can cost money

The intent is $0, and inside the free allowance it is $0 — but the allowance is what makes it free, not the configuration. What to watch:

- **Cloud Run** beyond the monthly free requests, vCPU-seconds and GiB-seconds, or beyond the free egress from North America. `min_instance_count` is 0 and CPU is allocated only during requests, so an idle service costs nothing; a service that is being used a great deal does not stay free.
- **Artifact Registry** beyond half a gigabyte of images. The cleanup policies keep the repository small; a deployment that pushes many images a day should check that they are working.
- **Secret Manager** beyond six active versions or the monthly free accesses. A version is read once per instance start, and the rotation keeps at most three versions live.
- **Cloud Logging** beyond the free monthly ingestion.
- **Cloud Trace** beyond 2.5 million spans a month. The platform's own traces of incoming requests are not billed at all; these are the spans the service adds inside them, and the sampler is what keeps their number a fraction of the requests.
- **Cloud Monitoring** beyond 150 MiB a month of ingested metrics of our own. The platform's own metrics of the service are free; ours are the three of section 12.5 of the spec. What passes that allowance is not traffic but series, and a series is created by every new combination of labels — which is why every label here is drawn from a closed list and no label ever carries anything a caller chose. One label holding a user or a task identifier would pass it in a week.
- **A region that is not first-tier**, where the free allowance does not apply.
- **Cloud DNS**, if the domain's records are ever moved into it: a managed zone is billed per month whether anybody visits or not. That is why DNS stays at the registrar and those two records are made by hand.
- **A load balancer**, if the domain mapping is ever replaced by one. Domain mappings cost nothing; a forwarding rule is billed by the hour.

The spend alert warns, it does not stop anything: Google has no switch that halts a project at a number. Treat the first alert as a fault to investigate.

**A caveat about the domain mapping.** Google labels Cloud Run domain mappings a preview feature and says they are not production-ready, citing latency. They are the only way to put a custom domain in front of Cloud Run without paying for a load balancer, which is why they are used here, and the alternative — a global external Application Load Balancer — is a monthly bill rather than a free tier.

## When the bootstrap changes

The bootstrap grants the identity that applies the configuration an enumerated set of roles, and that set grows with what the configuration describes. A deployment created before it grew does not learn of it: the script ran once, and nothing runs it again.

What that looks like is a delivery that fails at the first job, on something it had never done before — `terraform apply` refused with a 403 on a resource that is new in this change. Because the image is rolled out only after the cloud is brought up to date, nothing is deployed at all until it is fixed, and because a pull request only reads the configuration, the checks on it were green.

Run the bootstrap again. It is safe: every step checks before it creates, and the grants are idempotent.

```bash
TF_VAR_billing_account=01ABCD-234567-89EFGH just bootstrap
```

Then start the delivery again — from the branch, by hand, or by merging the next thing. Nothing was created halfway, so there is nothing to undo.

**The one that exists today.** The service sends its own traces and metrics, which needs two roles on the runtime identity — and those are the first bindings this configuration makes on the project itself, which the applying identity was not previously allowed to make. A deployment older than that change needs the single grant below, or the bootstrap above:

```bash
gcloud projects add-iam-policy-binding PROJECT_ID \
  --member="serviceAccount:mathtrail-tf@PROJECT_ID.iam.gserviceaccount.com" \
  --role=roles/resourcemanager.projectIamAdmin --condition=None
```

That the runtime identity ended up with what it needed is one command to confirm:

```bash
gcloud projects get-iam-policy PROJECT_ID \
  --flatten="bindings[].members" \
  --filter="bindings.members:mathtrail-run@PROJECT_ID.iam.gserviceaccount.com" \
  --format="value(bindings.role)"
```

It should name `roles/telemetry.writer` and `roles/serviceusage.serviceUsageConsumer`, and nothing else. The runtime identity's one other right — reading the two secrets — is granted on those secrets rather than on the project, so it does not appear here and its absence from this list is not a fault.

None of this is billed: granting a role and enabling an API cost nothing.

## Rotating the sealing key

Both keys are versions of the same secret. Add a version, then say in `prod.auto.tfvars` which version seals and which one is still accepted, and merge:

```bash
head -c 32 /dev/urandom | base64 | tr -d '\n' \
    | gcloud secrets versions add mathtrail-seal-key --data-file=-
```

```hcl
seal_key_version          = "3"
seal_key_previous_version = "2"
```

Tokens sealed with the previous key keep working until they expire on their own, so nobody is signed out by a rotation. Only once the deployment is serving with the new version does the version that fell off the end get disabled.

## When the pipeline is not available

Everything the pipeline does is a recipe, so the same delivery runs from a laptop — for an emergency, or to see what a step does:

```bash
export TF_VAR_billing_account=01ABCD-234567-89EFGH   # as in the repository secret
gcloud auth application-default login                # the credentials Terraform reads
just ci-tf-apply                                     # the cloud, as described
image=$(just ci-image-push "$(terraform -chdir=infra/terraform output -raw image_repository)/server")
just ci-deploy mathtrail us-central1 "$image"
just ci-smoke https://mcp.example.com
```

A rollback is one of these too, and it is a revision rather than an apply — Terraform does not know which digest is live, and Cloud Run keeps every revision that ever served:

```bash
gcloud run revisions list --service=mathtrail --region=us-central1
gcloud run services update-traffic mathtrail --region=us-central1 --to-revisions=REVISION=100
```

## Removing a deployment

`terraform destroy` removes everything the configuration created, including the secrets and every version in them: copy the sealing key out first if anything sealed with it still matters. The APIs stay enabled, because switching one off breaks whatever else in the project still uses it. The project, the state bucket, the federated pool, the DNS records and the OAuth client were never described here and stay as they are.
