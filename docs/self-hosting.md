# Running your own copy

MathTrail is MIT-licensed and holds nothing of its own: every deployment is one Cloud Run service, one Artifact Registry repository, three secrets and a domain, in a Google Cloud project you control, beside a small site on GitHub Pages that carries its privacy policy and terms. This page is how that gets created — and almost none of it is done by hand. The repository describes the deployment, and its workflows deliver it.

The name is not part of the licence: a copy that runs publicly goes by a name of its own ([below](#a-name-of-your-own)).

## What is done by hand, ever

Five things, once — four because no API can do them, and the settings of the repository itself — and a sixth for a copy that keeps the counts of how it is used:

1. **The bootstrap**, below: the project, the bucket its state lives in and the identity the pipeline signs in as. A pipeline cannot create the door it walks in by. Once — and again on the day that identity needs a right it was not given, which is its own section further down.
2. **The Google consent screen and one OAuth client.** Google has no API for either.
3. **The DNS records**, at whatever registrar holds the domain — for the service and for the site — and one click that makes the deployment identity a verified owner of the domain.
4. **The OAuth client's secret**, pasted into the repository's secrets. It is the one value that exists nowhere but the Google console.
5. **The repository's settings**: the fork's workflows enabled, GitHub Pages published from them on the site's domain, and that domain verified for Pages.
6. **The two reports of the counts**, in Data Studio (formerly Looker Studio), where `analytics` is on: Data Studio has no API that makes a report ([below](#the-counts-kept-for-years)).

Everything else — enabling APIs, the registry and its cleanup, the secrets, the service, the domain mapping, the spend alert, the counts kept for years, the image and every roll-out after it — happens when a change reaches the main branch, or when the workflow is started by hand from a branch. The site is published the same way, by a workflow of its own.

## What you need

- A Google Cloud project with billing attached, or the right to create one, and a billing account you may create budgets on.
- A domain, and the ability to add records to it. The service answers on a host of its own — `mcp.example.com` — and the site on another, the apex `example.com` for instance.
- A GitHub account to keep your copy of the repository in: its workflows check it, deliver it and publish the site.
- An address parents can write to. The privacy policy and the terms name whoever runs a copy, because that is who the parents using it are trusting.
- The development container of this repository, or a browser with Cloud Shell: the bootstrap is the only step that needs a shell at all.

## 1. Your copy of the repository

Fork this repository. GitHub runs none of a fork's workflows until they are enabled, on the fork's **Actions** tab: enable them there. From then on the fork is the description of your deployment, and every step below is either a change to it or a setting beside it.

What each workflow needs in a copy:

- **`deploy.yml`**, the delivery, needs everything below.
- **`pages.yml`**, the site, needs Pages published from GitHub Actions — step 4.
- **`ci.yml`**, the checks, needs nothing to run. Two of its steps report to this repository's projects: the SonarCloud job, which fails on `main` without a `SONAR_TOKEN` — give it a project of your own in `sonar-project.properties`, or remove it — and the Codecov upload, which fails nothing without a token of yours.
- **`release.yml`** needs green checks: it runs once `ci.yml` has passed on `main`, publishes a release and then runs `deploy.yml` with it — so a SonarCloud job left failing stops every release, and every delivery with it.
- **`codeql.yml`** and **`scorecard.yml`** need nothing.
- **`load.yml`** needs nothing, and spends runner minutes once a week; disable it if nobody reads its reports.

The checks and the site's build run inside the toolchain image this repository publishes, which anybody may pull. A copy that changes `.devcontainer/Dockerfile` publishes an image of its own from both workflows, so `TOOLCHAIN_IMAGE` in the `justfile` then has to name a registry the copy may write to.

### A name of your own

MathTrail is this project's name, and the licence does not pass it on: a copy that runs publicly goes by a name of its own. The name is in the words people read — the widget's dictionaries in `web/locales/`, the sign-in pages in `internal/transport/oauth/pages/`, what the tools tell the model in `internal/transport/mcp/` and `content/instructions/`, the server's name in the protocol, the folder and the file in a parent's Drive, and the site. `git grep -i mathtrail` finds every place, along with names nobody reads — the module path `github.com/MathTrail/…` in the Go imports, the names of the cloud resources — which can stay.

## 2. Say what the deployment is

Three files in the repository, and nothing about the deployment lives anywhere else:

| File | What it holds |
|---|---|
| `infra/terraform/prod.auto.tfvars` | project, region, host, OAuth client id, repository and its owner id, pool id — everything except the billing account |
| `infra/terraform/backend.hcl` | the bucket the state lives in |
| `infra/ci.env` | the two facts a workflow needs before it can sign in — filled from what the bootstrap prints |

Every value in them is this deployment's: replace each one with yours, the repository and its owner's id included. None of it is secret: the client id travels in every authorization request, and the federated pool is guarded by a condition on the repository rather than by the secrecy of its name. Changing where this repository deploys is a pull request, which is the point.

## 3. Bootstrap, once

```bash
TF_VAR_billing_account=01ABCD-234567-89EFGH just bootstrap
```

Or, with no local tooling at all, paste the contents of `infra/bootstrap.sh` into Cloud Shell. It is safe to run again: every step checks before it creates.

It creates the project and links billing, enables the handful of APIs without which nothing else can be created, makes the state bucket with versioning, creates the federated pool and its GitHub provider, and creates the identity the infrastructure is applied as — with an enumerated set of roles rather than Owner. It prints two lines for `infra/ci.env`. Commit them.

## 4. Your site: the privacy policy and the terms

Google publishes a consent screen only with a home page, a privacy policy and terms of use to link to, and the service's own consent page links the last two, at `/en/privacy/` and `/en/terms/` of the site. The ones in `site/` are this deployment's: they name its operator and the address that answers parents' questions. A copy publishes its own.

1. **The texts.** Rewrite `site/content/<locale>/privacy.md` and `terms.md` for your copy, in every language the site carries: the contact, the operator, and every sentence that stops being true of a copy. The English text is the one that prevails. Rewrite `about.yaml` too, since the page says who makes the service and yours is made by you, and `index.yaml`, the home page, which tells a parent how to add the service to a chat. The name the header shows, the link to the code, the address that answers questions, which the footer and the page "About" write out, and the connector's address the home page gives are in `web/src/site/brand.ts` — a test of the build holds the address that answers questions to the one the privacy policy names — and the footer's other words in `web/src/site/locales/`. The pictures a shared link to the site shows are its home page in each language, kept in `site/assets/`: `just site-og` photographs yours once the texts are. `just site-serve` shows the site on this machine before anybody else sees it.
2. **The address.** `SITE_BASE` in the `justfile` is the origin the site is published on, and every absolute address on the site is built from it. The service learns the same origin from `MATHTRAIL_SITE_URL`, [in `settings`](#the-services-environment).
3. **Pages.** In the repository's **Settings → Pages**, choose **GitHub Actions** as the source and enter your domain as the custom domain: a site published by a workflow is given its domain there, and the `CNAME` file the build writes is ignored. Verify the domain for Pages first, in your account's or organisation's settings, as GitHub recommends, so that no other repository can claim it.
4. **The DNS records.** For an apex domain, the four A and four AAAA records in [GitHub's list](https://docs.github.com/en/pages/configuring-a-custom-domain-for-your-github-pages-site/managing-a-custom-domain-for-your-github-pages-site); for a subdomain, a CNAME to `<owner>.github.io`. Once Pages offers it, turn on **Enforce HTTPS**: until then the policy is served over plain HTTP as well.

The site is published on every push to `main`. Have it live before the consent screen: Google asks for the policy's address.

The progress links each topic's name to its page on this site, and the words for the model name those pages too, at addresses built from `MATHTRAIL_SITE_URL`, the page's language and the topic's slug. Nothing turns the links off, so publish the site from the commit the service is delivered from: a service ahead of its site links to pages the site does not have yet.

## 5. The Google sign-in

The parent signs in with Google, and the service asks Google for exactly two things: a verified identifier, and room for one file of its own in the parent's Drive. It is configured in the Google Cloud console, under **Google Auth Platform**, on its four pages.

**Branding.** The name people see on the consent screen, a support email, a developer contact email, and three links: the home page, the privacy policy and the terms of your site, which for this deployment are `https://mathtrail.app/en/`, `https://mathtrail.app/en/privacy/` and `https://mathtrail.app/en/terms/`. Add the top private domain of those links — `mathtrail.app` here — as an authorized domain: one entry covers both the site on the apex and the service on its subdomain. Leave the logo out unless you mean to have the app reviewed: uploading one sends it to Google's brand verification.

**Audience.** User type **External**, publishing status **In production**. Not Testing: there only the test users listed on that page may sign in at all, and a consent expires seven days after it is given and takes the refresh token with it, so every parent would be signed out once a week.

**Data access.** Two scopes and no others: `openid`, which is what makes the identifier verified, and `https://www.googleapis.com/auth/drive.file`, which reaches only the files the app itself created plus anything the parent hands it explicitly. Both are non-sensitive, so publishing needs no app verification at all; brand verification is the separate, lighter process that makes the app's own name and logo appear on the consent screen instead of the project's name.

**Clients.** One client, type **Web application**, with a single authorized redirect URI — `https://<your host>/oauth/callback`, no trailing slash. No authorized JavaScript origins: the sign-in is a redirect the service performs, never a script inside a page. The client id goes into `prod.auto.tfvars`.

Neither the project nor this client is ever recreated. A parent grants `drive.file` to that client in that project, and the grant cannot be moved: a new client sees none of the files parents have already given it, and every child's profile becomes unreachable through the app.

## 6. The service's domain

1. Verify ownership of the domain in [Search Console](https://search.google.com/search-console), with the TXT record it asks for.
2. At the registrar: the service's host as a **CNAME** to `ghs.googlehosted.com.`
3. In Search Console, under the domain's settings, add `mathtrail-tf@<project>.iam.gserviceaccount.com` — the identity the configuration is applied as — as a verified owner. Cloud Run maps a domain only for an owner of it, and that identity is the one creating the mapping; without this the delivery gets as far as the domain and stops.

Do all three before the first delivery: the certificate is issued only once the record resolves, and that can take a day.

## 7. The two repository secrets

Settings → Secrets and variables → Actions → Secrets:

| Secret | What it is |
|---|---|
| `GOOGLE_OAUTH_CLIENT_SECRET` | The OAuth client's secret. Used exactly once, on the first apply of a deployment, to put the first version into Secret Manager; every apply after that leaves it alone. |
| `TF_VAR_billing_account` | The id of the billing account, of the form `01ABCD-234567-89EFGH`. It is the one fact about the deployment that is not written down beside the others, because it names the account that pays and this repository is public. Terraform reads variables from `TF_VAR_`-prefixed environment variables by itself. |

The bootstrap needs the second one too, so export it there: `TF_VAR_billing_account=01ABCD-234567-89EFGH just bootstrap`.

The key everything is sealed with is in neither list, and neither is the key the children are counted under in the log: the pipeline generates 32 random bytes for each on the first apply and writes them straight into Secret Manager. No person, no state file and no log ever sees either value.

## 8. Deliver

A push to `main` is delivered once the checks have passed on it and its release is published: `release.yml` hands the release to `deploy.yml`, which builds from the release's tag, so the service — `/health` and the header of the task card — names the release it runs. `deploy.yml` started by hand delivers any branch as it stands, named by how far past a release it is. Either way it runs three jobs:

1. **The cloud, as described.** `terraform apply`, every time, so the deployment always matches the commit it was cut from. On the very first run this is what creates everything; afterwards it is usually a no-op that takes half a minute.
2. **The image of this commit.** Built, pushed to the registry with the commit as its tag, and rolled out **by digest** — a tag can be moved afterwards and a digest cannot.
3. **The check.** `/health` is asked which commit it is serving, and the delivery fails unless the answer is the commit that was just built.

A delivery after a release that fails — an apply refused, a roll-out that timed out — is started again from its run with **Re-run failed jobs**. **Re-run all jobs** would run the release too, and publish the same commit once more under the next number.

A pull request that touches `infra/` gets a `terraform plan` in its summary instead, so what the cloud is about to become is reviewable before the merge.

**Give the telemetry twenty minutes.** On a project that has never received traces or metrics, enabling the APIs is all it takes, but the storage behind them is provisioned after the first data arrives rather than before. Until that finishes, reading a trace answers `_Trace bucket not found in project …` and the metrics are nowhere in Monitoring — which reads like a permanent fault and is not one. Nothing needs clicking; the service will have reported no error, because none of it failed.

## 9. Leave one entrance

Once the domain answers, withdraw the platform's own address for the service, so that there is one way in and one issuer of tokens: set `disable_default_url = true` in `prod.auto.tfvars` and merge. It is a step of its own because Cloud Run asks for the domain to be mapped before its own address is withdrawn — and until the certificate exists, that address is the only way to reach anything at all.

A deployment with no domain of its own instead sets `create_domain_mapping = false` and puts the platform's address in `public_host`, which is known only after the first delivery — so it takes two.

## 10. Connect it

Add the service to a chat the way the [README](../README.md#add-it-to-your-chat) describes, with your own address — `https://mcp.example.com/mcp` — and ask for a task. The first connection goes through your consent page, your Google client and your site, and the first task through every check: nothing else proves the whole path at once.

## The first delivery, as a checklist

- [ ] A fork with its workflows enabled, and a name of its own if it is to run publicly.
- [ ] `prod.auto.tfvars` and `backend.hcl` filled in, and a project id nobody has taken.
- [ ] `just bootstrap`, and the two lines it prints committed to `infra/ci.env`.
- [ ] Your privacy policy and terms published on your site — Pages from GitHub Actions, its domain verified, HTTPS enforced — which `SITE_BASE` and `MATHTRAIL_SITE_URL` both name.
- [ ] The consent screen published and a Web client created; its client id in `prod.auto.tfvars`.
- [ ] The domain verified, its CNAME created, and the deployment identity added as a verified owner.
- [ ] Both repository secrets: `GOOGLE_OAUTH_CLIENT_SECRET` and `TF_VAR_billing_account`.
- [ ] Merged to `main`, its checks and its release green, and all three jobs of the delivery after them.
- [ ] `disable_default_url = true` merged, once the domain answers.
- [ ] A task in a chat, through your own sign-in.

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
| `cpu`, `memory` | `1`, `1Gi` | Per instance, allocated only while serving. `cpu` is whole vCPUs — `1`, `2`, or `1000m`, `2000m` — and the sandbox runs a solver on each; `memory` is at least `1Gi` for every vCPU at the default step ceiling, and more in proportion to a higher one set in `settings`; nine tenths of it is the Go runtime's soft limit (`GOMEMLIMIT`) |
| `request_timeout` | `60s` | Before the platform cuts a request off |
| `disable_default_url` | `false` | Whether the platform's own address stops resolving; turned on once the domain answers |
| `create_domain_mapping` | `true` | Whether `public_host` is mapped onto the service |
| `seal_key_version` | `1` | The secret version the service seals with |
| `seal_key_previous_version` | empty | The version still accepted while a key is being rotated |
| `learner_key_version` | `1` | The secret version the children are counted under in the log; never rotated with the sealing key, and only at the turn of a month |
| `google_client_secret_version` | `1` | The secret version the sign-in authenticates with |
| `analytics` | `false` | Whether the counts of how the service is used are kept for years: the lines they are made from kept 62 days in a log bucket of their own, counted every night into BigQuery, and views for two reports ([below](#the-counts-kept-for-years)). Turn it on once your privacy policy says so |
| `settings` | `{}` | Extra environment variables of the service — the site's address, ceilings and timeouts, never a secret. The solver slots and `GOMEMLIMIT` follow `cpu` and `memory`, and are refused here |
| `budget_amount`, `budget_currency` | `1`, `USD` | Where the spend alert fires |
| `keep_images` | `5` | Image versions kept whatever their age |
| `image_max_age` | `30d` | When an older version is deleted |

### The service's environment

The service reads its configuration from environment variables alone; every one of them, with its default and what reads it, is in [section 11.2 of the spec](../SPEC.md#112-the-environment). A deployment described here sets them three ways:

- **From the variables above**, by Terraform: `MATHTRAIL_PUBLIC_URL` from `public_host`, `MATHTRAIL_GOOGLE_CLIENT_ID` from `google_oauth_client_id`, `MATHTRAIL_GCP_PROJECT_ID` from `project_id`, `MATHTRAIL_SOLVER_CONCURRENCY` from `cpu`, and `GOMEMLIMIT` from `memory`. `PORT` and `K_SERVICE` come from Cloud Run itself.
- **From Secret Manager**, read by Cloud Run as an instance starts: `MATHTRAIL_SEAL_KEY_CURRENT`, `MATHTRAIL_SEAL_KEY_PREVIOUS` during a rotation, `MATHTRAIL_LEARNER_KEY` and `MATHTRAIL_GOOGLE_CLIENT_SECRET`.
- **From the image**, which sets it: `MATHTRAIL_COUNTRY_DB`, the database of countries the image carries.
- **Through `settings`**, everything else. One of them a copy has to set, because its default is this deployment's site:

```hcl
settings = {
  MATHTRAIL_SITE_URL = "https://example.com"
}
```

The rest have the defaults the service is meant to run with. The ones a deployment is most likely to move:

| Variable | Default | What it moves |
|---|---|---|
| `MATHTRAIL_DAILY_TASKS` | `20` | Tasks a child may be given in a day |
| `MATHTRAIL_DAILY_FAILED` | `5` | Requests of a day that may end with the model out of attempts |
| `MATHTRAIL_RATE_USER_PER_MIN` | `60` | Requests one account may send an instance in a minute: a card waiting for a task asks fifteen times a minute |
| `MATHTRAIL_RATE_IP_PER_MIN` | `20` | Requests one address may send the sign-in of an instance in a minute |
| `MATHTRAIL_RATE_INSTANCE_PER_MIN` | `200` | Requests one instance takes in a minute, from everybody together |
| `MATHTRAIL_SOLVER_STEPS` | `25000000` | How far one solver may run; a higher ceiling needs more `memory` |
| `MATHTRAIL_TELEMETRY` | `auto` | Whether traces and metrics are sent: from a deployment, anywhere, or nowhere |
| `MATHTRAIL_LOG_LEVEL` | `info` | How much the service logs |

## What nothing in this repository owns

- **The project's billing.** Linked by the bootstrap, owned by whoever pays.
- **The values of the secrets.** Two are generated by the pipeline and read by nobody; the third is pasted once. None is ever in the state.
- **The months of the database of countries.** Each is published from a developer's machine with `just countries-publish`, as an image of its own; see below.
- **The Google OAuth client and the consent screen.** No API exists; they are the reason step 5 is done by hand.
- **The two reports of the counts.** Data Studio has no API that makes a report; they are made by hand, [below](#the-counts-kept-for-years).
- **Domain ownership and DNS.** At the registrar and in Search Console.
- **The site's Pages settings.** Its source and its domain, in the repository's settings.
- **The pool the pipeline signs in through.** Created by the bootstrap, deliberately outside Terraform.

## What can cost money

The intent is $0, and inside the free allowance it is $0 — but the allowance is what makes it free, not the configuration. What to watch:

- **Cloud Run** beyond the monthly free requests, vCPU-seconds and GiB-seconds, or beyond the free egress from North America. `min_instance_count` is 0 and CPU is allocated only during requests, so an idle service costs nothing; a service that is being used a great deal does not stay free.
- **Artifact Registry** beyond half a gigabyte of images. The cleanup policies keep the repository small; a deployment that pushes many images a day should check that they are working.
- **Secret Manager** beyond six active versions or the monthly free accesses. A version is read once per instance start, and the rotation keeps at most four versions live: three of the sealing key and the one key the children are counted under.
- **Cloud Logging** beyond the free monthly ingestion. With `analytics` on, the lines the children are counted from are taken in twice, into the default bucket and into their own, and kept there past 30 days at $0.01 a GiB a month: kilobytes a day.
- **BigQuery** beyond 1 TiB of queries and 10 GiB of storage a month, with `analytics` on. The nightly query reads megabytes and the tables hold kilobytes; a report reads them through its own cache. Google's price list names no charge for a scheduled query beyond its query, and the spend alert is what would tell otherwise.
- **Cloud Trace** beyond 2.5 million spans a month. The platform's own traces of incoming requests are not billed at all; these are the spans the service adds inside them, and the sampler is what keeps their number a fraction of the requests.
- **Cloud Monitoring** beyond 150 MiB a month of ingested metrics of our own. The platform's own metrics of the service are free; ours are the three of section 12.5 of the spec. What passes that allowance is not traffic but series, and a series is created by every new combination of labels — which is why every label here is drawn from a closed list and no label ever carries anything a caller chose. One label holding a user or a task identifier would pass it in a week.
- **A region that is not first-tier**, where the free allowance does not apply.
- **Cloud DNS**, if the domain's records are ever moved into it: a managed zone is billed per month whether anybody visits or not. That is why DNS stays at the registrar and those records are made by hand.
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

It should name `roles/telemetry.writer` and `roles/serviceusage.serviceUsageConsumer`, and nothing else. The runtime identity's one other right — reading the three secrets — is granted on those secrets rather than on the project, so it does not appear here and its absence from this list is not a fault.

**The two of the counts.** With `analytics` on, the configuration creates a log bucket, its sink and its link to BigQuery, and datasets, tables, views and a scheduled query, which need `roles/logging.configWriter` and `roles/bigquery.admin` on the applying identity. The bootstrap above grants both; a deployment that ran it before needs it again, or these two grants:

```bash
for role in roles/logging.configWriter roles/bigquery.admin; do
  gcloud projects add-iam-policy-binding PROJECT_ID \
    --member="serviceAccount:mathtrail-tf@PROJECT_ID.iam.gserviceaccount.com" \
    --role="$role" --condition=None
done
```

None of this is billed: granting a role and enabling an API cost nothing.

## Keeping your copy current

Merge this repository's `main` into your fork's like any other change: the delivery follows the merge, once its checks have passed and its release is published, and the content and the instructions for the model travel inside the binary, so there is nothing else to update. Where the merge meets what you made your own — the site's texts, the name — the conflict is yours to resolve; and a delivery that fails at its first job after a merge is the case [the section above](#when-the-bootstrap-changes) describes.

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

## The key the children are counted under

The lines a child is counted from name the child by a name that holds for a calendar month, derived from `mathtrail-learner-key` (section 12.1 of the spec). It is not rotated with the sealing key, and it is best never rotated: a new key gives every child a new name, and a month that sees both counts every child twice. If it has to be, do it at the turn of a month in UTC — add a version as for the sealing key, and set `learner_key_version` to it.

## The database of countries

The country a parent signs in from is looked up in DB-IP's IP to Country Lite, which DB-IP publishes once a month under the Creative Commons Attribution 4.0 licence and keeps to download for about three months. The runtime image copies one month of it from an image of this repository's own, pinned in the `Dockerfile` by tag and digest, so that a build never depends on how long DB-IP keeps a month. Publishing a newer month needs `docker`, and `gh` signed in with the right to write packages (`gh auth refresh -h github.com -s write:packages`). Nothing has to be signed in to GHCR by hand: the recipe signs `docker` in with the token `gh` holds, in a configuration of its own that it deletes when it ends, so the token is never written to `~/.docker`:

```bash
just countries-publish 2026-11   # without a month, this month's
just licenses
```

The recipe downloads the month from DB-IP, publishes it unchanged as `ghcr.io/mathtrail/mathtrail-standalone/dbip-country-lite:<month>`, never over a month already published, and pins the `Dockerfile` to it. The first time, the package is private: make it public in its settings on GitHub, or no build that is not signed in to GHCR can pull it. A copy of yours publishes under its own owner: change the name in the `justfile` and the `Dockerfile`. Every place that shows what the database found has to say it comes from DB-IP, with a link: `THIRD_PARTY_LICENSES` and the privacy policy do.

A binary run without the image needs the file itself: download a month from DB-IP, unzip it and name it in `MATHTRAIL_COUNTRY_DB`. A deployment refuses to start without one.

## The counts kept for years

A copy that has to show how much it is used — for a grant application, say — sets `analytics = true` in its tfvars, once its privacy policy says what is kept (this repository's does). The service writes the lines either way; what the switch adds is keeping their counts. Section 12.4 of the spec says what is counted and shown, and decision R192 of [decisions.md](decisions.md) why.

- **What the delivery creates.** A log bucket, `activity`, that keeps the lines the children are counted from for 62 days, and the sink that copies them into it; its link to BigQuery, `activity_logs`; the dataset `impact`, whose tables are the counts and the only thing kept for years; a scheduled query that fills them every night at half past one in UTC, run by an identity of its own that may read the lines, write the counts and nothing else; and two datasets of views, `impact_private` with exact numbers and `impact_public` with no group of fewer than ten children.
- **Before the first delivery with it**, run the bootstrap again: the applying identity needs two roles more ([below](#when-the-bootstrap-changes)). The first apply may stop at the scheduled query while the rights of its identity spread through the platform; start the delivery again.
- **A night that did not run** is made up for by the next: every night counts again every period the bucket still holds whole. A run started by hand in the console (BigQuery → Scheduled queries → Run now) does the same. Nothing tells you of a failure but the numbers themselves, and `just impact` lists the days no night counted.
- **The numbers.** `just impact` prints the main numbers of the latest months as the public views show them, and `just impact private` the exact ones. `just analytics-test` runs the SQL against a BigQuery emulator before it ever reaches the project.

The reports are made by hand, in [Data Studio](https://datastudio.google.com) (formerly Looker Studio), signed in as the project's owner:

1. **The private one.** Create a report and add data from BigQuery: the project, the dataset `impact_private`, one view at a time — `months`, `learners`, `learning`, `topics`, `traps`. Share it with nobody.
2. **The public one.** Create another from the views of `impact_public` — `months`, `learners_by_*`, `learning`, `topics`, `traps_by_topic`, `traps_by_grade`. In each data source, set *Data credentials* to the owner's, so that a viewer reads the views without a Google account of the project's. Write on the report, where the countries are shown, "IP geolocation by [DB-IP](https://db-ip.com)": the database the country of the sign-in comes from asks for it wherever what it found is shown. Share it as *Anyone on the internet with the link can view*.
3. **Link the public one** from the site and the `README.md`.

Leave the reports' data freshness at its default: a report then reads BigQuery a few times a day however many people open it.

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

`terraform destroy` removes everything the configuration created, including the secrets and every version in them: copy the sealing key out first if anything sealed with it still matters. The APIs stay enabled, because switching one off breaks whatever else in the project still uses it. The project, the state bucket, the federated pool, the DNS records and the OAuth client were never described here and stay as they are, and so does the site, until Pages is switched off in the repository's settings.

With `analytics` on, the destroy stops at the tables of the counts: they are years of evidence, and protected from deletion on purpose. To remove them too, set `deletion_protection = false` on them in `infra/terraform/analytics/main.tf`, apply, and destroy again. The log bucket goes, and its name stays taken for seven days, so `analytics` turned off and on again within a week fails until they pass. The Data Studio reports were never described here either.
