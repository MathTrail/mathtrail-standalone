# Running your own copy

MathTrail is MIT-licensed and holds nothing of its own: every deployment is one Cloud Run service, one Artifact Registry repository, five secrets — two of them empty until a copy lets a directory's reviewers in — and a domain, in a Google Cloud project you control, beside a small site on GitHub Pages that carries its privacy policy and terms. This page is how that gets created — and almost none of it is done by hand. The repository describes the deployment, and its workflows deliver it.

The name is not part of the licence: a copy that runs publicly goes by a name of its own ([below](#a-name-of-your-own)).

## What is done by hand, ever

Five things, once — four because no API can do them, and the settings of the repository itself — and three more for a copy that keeps the counts of how it is used:

1. **The bootstrap**, below: the project, the bucket its state lives in and the identity the pipeline signs in as. A pipeline cannot create the door it walks in by. Once — and again on the day that identity needs a right it was not given, which is its own section further down.
2. **The Google consent screen and one OAuth client.** Google has no API for either.
3. **The DNS records**, at whatever registrar holds the domain — for the service and for the site — and one click that makes the deployment identity a verified owner of the domain.
4. **The OAuth client's secret**, pasted into the repository's secrets. It is the one value that exists nowhere but the Google console.
5. **The repository's settings**: the fork's workflows enabled, GitHub Pages published from them on the site's domain, and that domain verified for Pages.
6. **The two reports of the counts**, in Data Studio (formerly Looker Studio), where `analytics` is on: Data Studio has no API that makes a report ([below](#the-counts-kept-for-years)).
7. **The GitHub App that brings the live numbers to the site**, with its two environments, where `analytics` is on: GitHub makes an App only for a person on its pages ([below](#the-counts-kept-for-years)).
8. **The daily report**, where `analytics` is on: the export of the costs turned on in the Billing console, and the report's page and its schedule in Data Studio. Neither has an API ([below](#the-daily-report)).

Everything else — enabling APIs, the registry and its cleanup, the secrets, the service, the domain mapping, the spend alert, the load alert, the counts kept for years, the image and every roll-out after it — happens when a change reaches the main branch, or when the workflow is started by hand from a branch. The site is published the same way, by a workflow of its own.

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
- **`pages.yml`**, the site, needs Pages published from GitHub Actions — step 4. `release.yml` runs it with every release. A site published from it by hand makes its own data and offers no PDF of the paper. A pull request's site is built and checked by `research.yml`, and never published.
- **`ci.yml`**, the checks, needs nothing to run. Two of its steps report to this repository's projects: the SonarCloud job, which fails on `main` without a `SONAR_TOKEN` — give it a project of your own in `sonar-project.properties`, or remove it — and the Codecov upload, which fails nothing without a token of yours.
- **`release.yml`** needs green checks: it runs once `ci.yml` has passed on `main`, publishes a release and then runs `deploy.yml` with it, and beside it `research.yml` and then `pages.yml` — so a SonarCloud job left failing stops every release, and every delivery and every publication of the site with it. Once the release is delivered it photographs the release's cards and writes the pictures over the branch `screens`; the README shows this repository's pictures, since its addresses name it, until you point them at your own, there and in `published` in `web/scripts/screens.ts`, which its test holds the README to.
- **`research.yml`** needs nothing. On every pull request and every release it holds the experiments of paper A to the results kept in `research/`, builds the paper, and makes the data of the page "Research" with the paper's PDF, whatever placeholder it still prints. `release.yml` runs it with every release, before `pages.yml`, which builds the site from them, so a release whose experiments no longer give their results, or whose paper fails, publishes no site. On a pull request it builds the site from them itself, and publishes nothing.
- **`paper.yml`** needs nothing. After every release whose paper A or its artifact changed, it publishes a release of the paper, tagged `paper-a/…`, with the artifact named and anonymous, each shown first to reproduce the experiments from itself, and the paper's PDFs. The paper is this deployment's authors' work, so a copy disables it, as it takes out the paper's title and PDF ([below](#4-your-site-the-privacy-policy-and-the-terms)).
- **`zenodo.yml`** runs only when started by hand, and puts a release of the paper on Zenodo: it needs an environment `zenodo` that only `main` may run in, with a Zenodo token as its secret `ZENODO_TOKEN`. A copy disables it with `paper.yml`.
- **`codeql.yml`** and **`scorecard.yml`** need nothing.
- **`load.yml`** needs nothing, and spends runner minutes once a week; disable it if nobody reads its reports.
- **`live.yml`** brings the live numbers of the page "Research" to `main` once a month. It needs the counts kept for years, a GitHub App of your own and two environments ([below](#the-counts-kept-for-years)), and fails every month without them: disable it until your copy has them.

The checks and the site's build run inside the toolchain image this repository publishes, which anybody may pull. A copy that changes `.devcontainer/Dockerfile` publishes an image of its own from both workflows, so `TOOLCHAIN_IMAGE` in the `justfile` then has to name a registry the copy may write to.

### A name of your own

MathTrail is this project's name, and the licence does not pass it on: a copy that runs publicly goes by a name of its own. The name is in the words people read — the widget's dictionaries in `web/locales/`, the sign-in pages in `internal/transport/oauth/pages/`, what the tools tell the model in `internal/transport/mcp/` and `content/instructions/`, the server's name in the protocol, the folder and the file in a parent's Drive, and the site. `git grep -i mathtrail` finds every place, along with names nobody reads — the module path `github.com/MathTrail/…` in the Go imports, the names of the cloud resources — which can stay.

What this deployment gives the chats' directories is under its name too: the texts of its listings in `docs/listing.md`, the pictures of its listing in Claude's directory, which `just listing-shots` takes from the preview, and the package for ChatGPT's directory in `plugin/`, with its icons. A copy that lists itself makes its own. One that does not removes the folder, with the recipes `plugin-icons`, `plugin-zip` and `listing-shots`, the scripts `web/scripts/icons.ts` and `web/scripts/listing.ts` with their tests, and the lines that name the folder, the scripts or their pictures in `.dockerignore`, `.gitignore`, `web/vitest.config.ts` and `sonar-project.properties`.

## 2. Say what the deployment is

Three files in the repository, and nothing about the deployment lives anywhere else:

| File | What it holds |
|---|---|
| `infra/terraform/prod.auto.tfvars` | project, region, host, OAuth client id, repository and its owner id, pool id — everything except the billing account |
| `infra/terraform/backend.hcl` | the bucket the state lives in |
| `infra/ci.env` | the facts a workflow needs before it can sign in — two from what the bootstrap prints, and, where the counts are kept, the identity that reads the live numbers |

Every value in them is this deployment's: replace each one with yours, the repository and its owner's id included. None of it is secret: the client id travels in every authorization request, and the federated pool is guarded by a condition on the repository rather than by the secrecy of its name. Changing where this repository deploys is a pull request, which is the point.

## 3. Bootstrap, once

```bash
TF_VAR_billing_account=01ABCD-234567-89EFGH just bootstrap
```

Or, with no local tooling at all, paste the contents of `infra/bootstrap.sh` into Cloud Shell. It is safe to run again: every step checks before it creates.

It creates the project and links billing, enables the handful of APIs without which nothing else can be created, makes the state bucket with versioning, creates the federated pool and its GitHub provider, and creates the identity the infrastructure is applied as — with an enumerated set of roles rather than Owner. It prints two lines for `infra/ci.env`. Commit them.

## 4. Your site: the privacy policy and the terms

Google publishes a consent screen only with a home page, a privacy policy and terms of use to link to, and the service's own consent page links the last two, at `/en/privacy/` and `/en/terms/` of the site. The ones in `site/` are this deployment's: they name its operator and the address that answers parents' questions. A copy publishes its own.

1. **The texts.** Rewrite `site/content/<locale>/privacy.md`, `terms.md` and `help.md` for your copy, in every language the site carries: the contact, the operator, and every sentence that stops being true of a copy. The help names the connector's address and this repository's issues and security advisories too. The English text is the one that prevails. Rewrite `about.yaml` too, since the page says who makes the service and yours is made by you, and `index.yaml`, the home page, which tells a parent how to add the service to a chat. The photographs on the page "About" are the family of this deployment's author, which the MIT License does not cover: put your own in their place in `site/assets/photos/`, which `just site-photos` makes from your originals under the names `web/src/site/brand.ts` gives them. Who is on the page is the list `familyMembers` in `brand.ts`, each with words under their name in every `about.yaml` and a portrait of their own. To show no photographs, take them off `web/src/site/AboutPage.tsx` and their `photo` texts out of every `about.yaml`, then remove `photos` and what names it from `brand.ts`, `site/assets/photos/`, `web/scripts/photos.ts` with its test, and the lines of the build's test that name the photographs. The page "Coach" shows a product of this deployment's author, not of yours: remove it with its prototype and the screen of it. That is `coach.yaml` in every language; `web/src/site/CoachPage.tsx` with its test; the page in `web/src/site/pages.ts`, its entry in the menu in `web/src/site/frame.ts` and its word, `nav.coach`, in `web/src/site/locales/`; `coachPrototypePath`, `coachScreenPath` and `coachScreenSize` in `web/src/site/brand.ts`, and the line of `robots` in `web/src/site/metadata.ts` that names it, with its test; the prototype, its fonts' licence and its screen in `site/assets/`, and their lines in `web/scripts/prerender-site.ts` and in its test, which also names the page; the page's addresses in `web/tools/sitecheck/published.ts`; the prototype's lines in `.gitattributes`, `.github/workflows/codeql.yml` and `sonar-project.properties`; and the prototype's section of the license list, in the `justfile`'s `_license-list`, before `just licenses` writes the list again. The page "Research" names the paper this deployment's authors wrote about MathTrail and offers its PDF once there is one: the paper is their work, so take out its title and its PDF. The title is `hero.title` and `hero.subtitle` in every `research.yaml`: give the page a heading of your own there, and take out of `web/src/site/ResearchPage.test.tsx` the test that holds them to the paper's sources. The PDF is built by `research.yml`, whose job "Paper A" hands it to the page's data once the paper prints no placeholder: take out the two steps that pass it on, "Hand the named paper on to the page's data" in that job and "Take the named paper the site offers" in the job "The page's data", and the page offers none. The rest of the page says what the service does, and its numbers are your copy's own, computed by `just research-data` from the code being built. Its live numbers are this deployment's children's, from the snapshot `site/research/live.json`: remove it, and the page says they are still to come until your copy's own `live.yml` brings one ([below](#the-counts-kept-for-years)). To show no page "Research" at all, remove `research.yaml` in every language; `web/src/site/ResearchPage.tsx`, `ResearchTheses.tsx`, `ResearchModel.tsx`, `ResearchLive.tsx`, `ResearchNumbers.tsx` and `research.ts`, with their tests and `web/src/site/testing/handtyped.ts`; `site/research/live.json`; the recipe `site-live` that takes it, with `_live-read` and `_live-check`, which it runs, and the workflow `live.yml`, which runs them too; the snapshot's SQL in `infra/analytics/site/` and its table and reader in `infra/terraform/analytics/snapshot.tf`, with their tests and what names them in the Terraform configuration and `infra/ci.env`; the page in `web/src/site/pages.ts`, its entries in the menu and the footer in `web/src/site/frame.ts`, and its words, `nav.research` and the `research.*` counts, in `web/src/site/locales/`; `research` in `site/data.json` and what reads it in `web/src/site/data.ts`; the build's `--research` and the paper's copy in `web/scripts/prerender-site.ts`, with the lines of its test and of `web/scripts/og.test.ts` that give them; `research-data` among what `site` makes first in the `justfile`; the job "The page's data" of `.github/workflows/research.yml`, with the steps of `pages.yml` and `research.yml` that make the data or take it, `data` among what the job "Build the site" of `research.yml` needs, and the steps of `.github/actions/site` that take the data and hold it to its commit; and the page's addresses in `web/tools/sitecheck/published.ts`. The name the header shows, the link to the code, the address that answers questions, which the footer and the page "About" write out, and the connector's address the home page gives are in `web/src/site/brand.ts` — a test of the build holds the address that answers questions to the one the privacy policy names — and the footer's other words in `web/src/site/locales/`. The pictures a shared link to the site shows are its home page in each language, kept in `site/assets/`: `just site-og` photographs yours once the texts are. `just site-serve` shows the site on this machine before anybody else sees it.
2. **The address.** `SITE_BASE` in the `justfile` is the origin the site is published on, and every absolute address on the site is built from it. The service learns the same origin from `MATHTRAIL_SITE_URL`, [in `settings`](#the-services-environment).
3. **Pages.** In the repository's **Settings → Pages**, choose **GitHub Actions** as the source and enter your domain as the custom domain: a site published by a workflow is given its domain there, and the `CNAME` file the build writes is ignored. Verify the domain for Pages first, in your account's or organisation's settings, as GitHub recommends, so that no other repository can claim it.
4. **The DNS records.** For an apex domain, the four A and four AAAA records in [GitHub's list](https://docs.github.com/en/pages/configuring-a-custom-domain-for-your-github-pages-site/managing-a-custom-domain-for-your-github-pages-site); for a subdomain, a CNAME to `<owner>.github.io`. Once Pages offers it, turn on **Enforce HTTPS**: until then the policy is served over plain HTTP as well.

The site is published with every release, from the release's tag, beside the delivery of the service, and the widget's cards on its pages name that release. Have it live before the consent screen, since Google asks for the policy's address: before the first release exists, start **Pages** by hand on `main`, which publishes the branch as it stands and names no release.

The progress links each topic's name to its page on this site, and the words for the model name those pages too, at addresses built from `MATHTRAIL_SITE_URL`, the page's language and the topic's slug. Nothing turns the links off, so publish the site from the commit the service is delivered from: a service ahead of its site links to pages the site does not have yet.

## 5. The Google sign-in

The parent signs in with Google, and the service asks Google for exactly one thing: room for one file of its own in the parent's Drive. It then asks Drive whose Drive the access token of that permission reaches, which is how it knows who signed in. It is configured in the Google Cloud console, under **Google Auth Platform**, on its four pages.

**Branding.** The name people see on the consent screen, a support email, a developer contact email, and three links: the home page, the privacy policy and the terms of your site, which for this deployment are `https://mathtrail.app/`, `https://mathtrail.app/en/privacy/` and `https://mathtrail.app/en/terms/`. Add the top private domain of those links — `mathtrail.app` here — as an authorized domain: one entry covers both the site on the apex and the service on its subdomain. Leave the logo out unless you mean to have the app reviewed: uploading one sends it to Google's brand verification.

**Audience.** User type **External**, publishing status **In production**. Not Testing: there only the test users listed on that page may sign in at all, and a consent expires seven days after it is given and takes the refresh token with it, so every parent would be signed out once a week.

**Data access.** One scope and no other: `https://www.googleapis.com/auth/drive.file`, which reaches only the files the app itself created plus anything the parent hands it explicitly. Not `openid` beside it: once a request asks for a sign-in too, Google shows the Drive with a box it leaves unticked, and a parent who presses on past it grants the sign-in alone. The scope is non-sensitive, so publishing needs no app verification at all; brand verification is the separate, lighter process that makes the app's own name and logo appear on the consent screen instead of the project's name.

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

A push to `main` is delivered once the checks have passed on it and its release is published: `release.yml` hands the release to `deploy.yml`, which builds from the release's tag, so the service — `/health` and the header of the task card — names the release it runs; beside it, `pages.yml` publishes the site from the same tag. `deploy.yml` started by hand delivers any branch as it stands, named by how far past a release it is. Either way it runs three jobs:

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

Add the service to a chat the way the [help page](https://mathtrail.app/en/help/) describes, with your own address — `https://mcp.example.com/mcp` — and ask for a task. The first connection goes through your consent page, your Google client and your site, and the first task through every check: nothing else proves the whole path at once.

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

It is not finished until all four of these say so, and the fifth too where `operator_email` names an operator:

```bash
# the domain answers, and answers as the commit that was deployed
just ci-smoke https://mcp.example.com

# the same version and commit in the line the service logs as it starts
gcloud run services logs read mathtrail --region=us-central1 --limit=20

# the spend alert exists and names this project
gcloud billing budgets list --billing-account=01ABCD-234567-89EFGH

# both cleanup policies are in force, and cleanupPolicyDryRun is false
gcloud artifacts repositories describe mathtrail --location=us-central1

# the load alert exists: mathtrail load
gcloud monitoring policies list --project=PROJECT_ID --format="value(displayName)"
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
| `github_oidc_subject_prefix` | empty | Where `analytics` is on: how GitHub begins the subject of the repository's tokens, which `gh api repos/OWNER/NAME/actions/oidc/customization/sub --jq .sub_claim_prefix` prints. The identity that reads the live numbers is lent to the jobs of one environment by the subject they are given ([below](#the-counts-kept-for-years)) |
| `workload_identity_pool_id` | `mathtrail-github` | The pool created by the bootstrap, in which a workflow's token is exchanged |
| `region` | `us-central1` | Has to offer domain mappings and be a first-tier region |
| `service_name` | `mathtrail` | Names the service, its images, its secrets and its identities |
| `image` | a placeholder | What a service created from nothing starts with; after that the delivery owns the field |
| `max_instances` | `5` | The ceiling on instances running at once |
| `concurrency` | `80` | Requests one instance serves at a time |
| `cpu`, `memory` | `1`, `1Gi` | Per instance, allocated only while serving. `cpu` is whole vCPUs — `1`, `2`, or `1000m`, `2000m` — and the sandbox runs a solver on each; `memory` is at least `1Gi` for every vCPU at the default step ceiling, and more in proportion to a higher one set in `settings`; nine tenths of it is the Go runtime's soft limit (`GOMEMLIMIT`) |
| `request_timeout` | `60s` | Before the platform cuts a request off |
| `disable_default_url` | `false` | Whether the platform's own address stops resolving; turned on once the domain answers |
| `create_domain_mapping` | `true` | Whether `public_host` is mapped onto the service |
| `seal_key_version` | `1` | The secret version the service seals with |
| `seal_key_previous_version` | empty | The version still accepted while a key is being rotated |
| `learner_key_version` | `1` | The secret version the children are counted under in the log; never rotated with the sealing key, and only at the turn of a month |
| `google_client_secret_version` | `1` | The secret version the sign-in authenticates with |
| `reviewer_password_version`, `reviewer_grant_version` | empty | The secret versions of the reviewers' sign-in, set together or not at all; empty, the service is given neither and the sign-in is off ([below](#the-reviewers-sign-in)) |
| `analytics` | `false` | Whether the counts of how the service is used are kept for years: the lines they are made from kept 62 days in a log bucket of their own, counted every night into BigQuery, and views for two reports ([below](#the-counts-kept-for-years)). Turn it on once your privacy policy says so |
| `billing_export` | `false` | Where `analytics` is on: whether the export of the costs has made its table in the dataset `billing`, so that the daily report's view can be made over it; set once the table is there ([below](#the-daily-report)) |
| `settings` | `{}` | Extra environment variables of the service — the site's address, ceilings and timeouts, never a secret. The solver slots and `GOMEMLIMIT` follow `cpu` and `memory`, and are refused here |
| `budget_amount`, `budget_currency` | `1`, `USD` | Where the spend alert fires |
| `operator_email` | empty | The address the load alert is mailed to; empty, there is no alert. It is written in the repository, so it has to be an address already public ([below](#the-load-alert)) |
| `load_alert_per_minute` | `300` | Requests a minute, on average over five minutes, past which the load alert is mailed |
| `keep_images` | `5` | Image versions kept whatever their age |
| `image_max_age` | `30d` | When an older version is deleted |

### The service's environment

The service reads its configuration from environment variables alone; every one of them, with its default and what reads it, is in [section 11.2 of the spec](../SPEC.md#112-the-environment). A deployment described here sets them three ways:

- **From the variables above**, by Terraform: `MATHTRAIL_PUBLIC_URL` from `public_host`, `MATHTRAIL_GOOGLE_CLIENT_ID` from `google_oauth_client_id`, `MATHTRAIL_GCP_PROJECT_ID` from `project_id`, `MATHTRAIL_SOLVER_CONCURRENCY` from `cpu`, and `GOMEMLIMIT` from `memory`. `PORT` and `K_SERVICE` come from Cloud Run itself.
- **From Secret Manager**, read by Cloud Run as an instance starts: `MATHTRAIL_SEAL_KEY_CURRENT`, `MATHTRAIL_SEAL_KEY_PREVIOUS` during a rotation, `MATHTRAIL_LEARNER_KEY` and `MATHTRAIL_GOOGLE_CLIENT_SECRET`, and `MATHTRAIL_REVIEWER_PASSWORD` with `MATHTRAIL_REVIEWER_GRANT` once their versions are named.
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
| `MATHTRAIL_OPENAI_CHALLENGE` | empty | The token ChatGPT's plugin directory gives when you submit your copy, which the service then serves, as it is, at `/.well-known/openai-apps-challenge` on its own domain; empty, that address is not served. At most 512 characters of visible ASCII. It proves the domain is yours, and is not a secret |

## What nothing in this repository owns

- **The project's billing.** Linked by the bootstrap, owned by whoever pays.
- **The values of the secrets.** Two are generated by the pipeline and read by nobody; the third is pasted once; the reviewers' two are made by two recipes, when a copy wants them. None is ever in the state.
- **The months of the database of countries.** Each is published from a developer's machine with `just countries-publish`, as an image of its own; see below.
- **The Google OAuth client and the consent screen.** No API exists; they are the reason step 5 is done by hand.
- **The two reports of the counts.** Data Studio has no API that makes a report; they are made by hand, [below](#the-counts-kept-for-years).
- **The App that brings the live numbers to the site, its key and its two environments.** Made by hand, [below](#the-counts-kept-for-years); the key is a secret of one environment alone.
- **The export of the costs, and the daily report's page and its schedule.** Only the Billing console turns the export on, and Data Studio has no API that makes a page or a schedule; both are done by hand, [below](#the-daily-report).
- **Domain ownership and DNS.** At the registrar and in Search Console.
- **The site's Pages settings.** Its source and its domain, in the repository's settings.
- **The pool the pipeline signs in through.** Created by the bootstrap, deliberately outside Terraform.

## What can cost money

The intent is $0, and inside the free allowance it is $0 — but the allowance is what makes it free, not the configuration. What to watch:

- **Cloud Run** beyond the monthly free requests, vCPU-seconds and GiB-seconds, or beyond the free egress from North America. `min_instance_count` is 0 and CPU is allocated only during requests, so an idle service costs nothing; a service that is being used a great deal does not stay free.
- **Artifact Registry** beyond half a gigabyte of images. The cleanup policies keep the repository small; a deployment that pushes many images a day should check that they are working.
- **Secret Manager** beyond six active versions or the monthly free accesses. A version is read once per instance start, and the rotation keeps at most four versions live: three of the sealing key and the one key the children are counted under. The reviewers' sign-in adds two while it is on, and each new password or grant one more until the version it replaces is destroyed.
- **Cloud Logging** beyond the free monthly ingestion. With `analytics` on, the lines the children are counted from are taken in twice, into the default bucket and into their own, and kept there past 30 days at $0.01 a GiB a month: kilobytes a day.
- **BigQuery** beyond 1 TiB of queries and 10 GiB of storage a month, with `analytics` on. The nightly query reads megabytes and the tables hold kilobytes; a report reads them through its own cache. The export of the costs, where it is on, writes kilobytes a day. Google's price list names no charge for a scheduled query beyond its query, and the spend alert is what would tell otherwise.
- **Cloud Trace** beyond 2.5 million spans a month. The platform's own traces of incoming requests are not billed at all; these are the spans the service adds inside them, and the sampler is what keeps their number a fraction of the requests.
- **Cloud Monitoring** beyond 150 MiB a month of ingested metrics of our own. The platform's own metrics of the service are free; ours are the three of section 12.5 of the spec. What passes that allowance is not traffic but series, and a series is created by every new combination of labels — which is why every label here is drawn from a closed list and no label ever carries anything a caller chose. One label holding a user or a task identifier would pass it in a week.
- **Cloud Monitoring's alerting**, where `operator_email` names an operator. Google's price list, read on 2026-10-08, starts charging for alerting no sooner than 1 September 2027, and promises notice 90 and 30 days before: $0.35 a month for each metric an alerting policy names, and $0.50 for every million points its queries return. For the one load alert that is some $0.39 a month, against a budget of $1. Its mail costs nothing.
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

**The two of the load alert.** Where `operator_email` names an operator, the configuration creates a notification channel and the alert that mails it, which need `roles/monitoring.notificationChannelEditor` and `roles/monitoring.alertPolicyEditor` on the applying identity. A pull request's checks do not call Cloud Monitoring and pass without them, so grant them before merging the change that brings the alert, or the delivery after it stops at the alert:

```bash
for role in roles/monitoring.notificationChannelEditor roles/monitoring.alertPolicyEditor; do
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

## The reviewers' sign-in

A directory that reviews a copy before listing it asks for an account its reviewers can sign in with at once, with a profile already filled. Google may ask a reviewer signing in from another country or device to confirm it is them, and ChatGPT's review turns away an app whose reviewer meets such a step. So the consent screen can offer a directory's reviewer, folded under the parent's way in, a password in place of Google, which signs them in as one demo account whose grant at Google the service holds (decision R222). It is off unless both of its secrets are set. It reaches nothing but the demo account's own file, and the demo account's lines count no child, so nothing a reviewer does reaches the counts kept for years.

**Turning it on.**

1. Make a Google account for the demo profile, and keep its 2-Step Verification on: only you ever sign in to it at Google. Check that the app is **In production** (section 5): an app in Testing loses every refresh token after seven days.
2. Deliver a commit that has the two secrets in it, so that they exist, empty.
3. Capture the demo account's grant, in a terminal of the devcontainer:

   ```bash
   just reviewer-grant
   ```

   Open the address it prints in a browser signed in as the demo account, and allow both permissions. Google sends the browser back to the service, which cannot read the request and says the link is not right: paste the address of that page into the terminal. The recipe exchanges the code with the client secret from Secret Manager and adds the account's identifier and its refresh token as a version of `mathtrail-reviewer-grant`, printing nothing of either. Secret Manager says which version it made.
4. Make the password:

   ```bash
   just reviewer-password
   ```

   It is 32 random characters, added as a version of `mathtrail-reviewer-password` and shown once: keep it in a password manager, and give it to the directories in their review forms, never anywhere else.
5. Name both versions in `prod.auto.tfvars`, and merge:

   ```hcl
   reviewer_password_version = "1"
   reviewer_grant_version    = "1"
   ```

   The service's start says `reviewer sign-in` with `configured` true, and the consent screen shows "Reviewing MathTrail for a directory?".

Fill the demo profile by signing in through that line yourself, in a chat, before the review. Its 20 tasks a day are shared by every reviewer, and the day starts at 00:00 UTC, so do not fill it on the day a review may begin. The progress shows a change over the past week only while the last answer is at most six days old, so answer a few tasks in the days before a review. Sign in through the line once yourself before a review begins as well: a grant that has ended shows only when somebody signs in.

**Never connect the demo account through Google.** A parent's account is known by its identifier at Drive, and the demo account by the identifier at Google its grant names, so the demo account signed in with Google is a parent's account like any other: its lessons are counted, and disconnecting that chat ends every grant of the demo account at MathTrail, the reviewers' among them. Fill and check the demo profile through the reviewers' line alone.

**Changing the password.** Run `just reviewer-password` again, name the new version and merge. A reviewer signed in before keeps their session to its own end, up to 90 days.

**Destroying an old version** of either secret waits until the deployment reads the new one. Until then, the revision serving names the old version, and Cloud Run reads a secret as an instance starts. With that version destroyed, no new instance can start, and the service goes down once the running ones stop. A revision from before the change names it too, so the service can no longer be rolled back to one. If it was destroyed first, put the new version in service at once:

```bash
gcloud run services update mathtrail --region=us-central1 \
  --update-secrets=MATHTRAIL_REVIEWER_PASSWORD=mathtrail-reviewer-password:N
```

For the grant, the pair is `MATHTRAIL_REVIEWER_GRANT=mathtrail-reviewer-grant:N`. Then merge the new version before anything else is deployed: a deployment from a commit that still names the old version brings the failure back.

**Cutting every reviewer off at once.** Remove MathTrail's access in the demo account's Google account, under the apps connected to it. Every session ends at its next renewal, within the hour, and the sign-in fails until the grant is captured again with `just reviewer-grant`.

**When the sign-in fails.** Google stops honouring the grant when it goes unused for six months, when MathTrail's access is removed in the demo account, or once the account has a hundred newer refresh tokens for the app. The reviewer's chat then fails to connect, and the line `auth_consent` says `failed` with `grant_ended`, at error level. Capture the grant again, name the new version, and merge. On 2026-10-06 the grant also stopped within hours of the demo account's password being changed, with MathTrail still among the account's connected apps and no other cause known ([the live report](live/14-reviewer-access.md)). Google documents that for Gmail's scopes alone, but capture the grant again whenever that password changes.

**Rotating the sealing key** changes nothing here: while the previous key is kept, the demo account is known by the identifiers of both keys, so a reviewer's session begun before the rotation is still the demo account's.

**Turning it off.** Empty both versions in `prod.auto.tfvars` and merge; the consent screen no longer offers the line.

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

- **What the delivery creates.** A log bucket, `activity`, that keeps the lines the children are counted from for 62 days, and the sink that copies them into it; its link to BigQuery, `activity_logs`; the dataset `impact`, whose tables are the counts and the only thing kept for years; a scheduled query that fills them every night at half past one in UTC, run by an identity of its own that may read the lines, write the counts and nothing else; and two datasets of views, `impact_private` with exact numbers and `impact_public` with no group of fewer than ten children. For the page "Research", a second scheduled query, run every day at half past two by the same identity, keeps the snapshot of the live numbers, one row of the public views, in the table `live` of a dataset of its own, `impact_site`; and an identity, `<service_name>-live`, may read that dataset and nothing else, and only a job of the GitHub environment `live-numbers` may borrow it. For the daily report, an empty dataset, `billing`, which the export of the costs writes into once it is turned on, and the view `impact_private.daily_report` over it once `billing_export` says so ([below](#the-daily-report)).
- **Before the first delivery with it**, run the bootstrap again: the applying identity needs two roles more ([below](#when-the-bootstrap-changes)). The first apply may stop at a scheduled query while the rights of its identity spread through the platform; start the delivery again.
- **A night that did not run** is made up for by the next: every night counts again every period the bucket still holds whole. A run started by hand in the console (BigQuery → Scheduled queries → Run now) does the same. Nothing tells you of a failure but the numbers themselves, and `just impact` lists the days no night counted.
- **The numbers.** `just impact` prints the main numbers of the latest months as the public views show them, and `just impact private` the exact ones. `just analytics-test` runs the SQL against a BigQuery emulator before it ever reaches the project.
- **The page "Research"** shows the latest month counted whole of how the answers came out against the chance promised, from the snapshot committed as `site/research/live.json`; the site's build reads no BigQuery. The workflow `live.yml` brings it from the second day of each month, when the night has counted the month before, running every day to the fifth so that a failed run or a late night is taken up. It reads the snapshot as `<service_name>-live`, holds it to the rule of the public views with the bench and, when it differs from the one `main` holds, has a GitHub App of your own open a pull request with it, which merges itself once the required checks pass; the release after the merge publishes the site. It runs by hand too, and, with *deliver* unticked, only shows what it read. `just site-live` reads the same table with your own credentials and writes the file, to look at before it is delivered. GitHub mails a failed run to whoever caused it, and the checks of the App's pull request and the runs after its merge are the App's: watch the repository's pull requests and releases.

The reports are made by hand, in [Data Studio](https://datastudio.google.com) (formerly Looker Studio), signed in as the project's owner:

1. **The private one.** Create a report and add data from BigQuery: the project, the dataset `impact_private`, one view at a time — `daily`, `weeks`, `months`, `learners`, `dose`, `learning`, `topics`, `traps`, `chances`, `kept_up`. Share it with nobody.
2. **The public one.** Create another from the views of `impact_public` — `months`, `learners_by_*`, `learning`, `topics`, `traps_by_topic`, `traps_by_grade`, `chances`, `chances_total`, `kept_up`. In each data source, set *Data credentials* to the owner's, so that a viewer reads the views without a Google account of the project's. Write on the report, where the countries are shown, "IP geolocation by [DB-IP](https://db-ip.com)": the database the country of the sign-in comes from asks for it wherever what it found is shown. Share it as *Anyone on the internet with the link can view*.
3. **Link the public one** from the site and the `README.md`.

Leave the reports' data freshness at its default: a report then reads BigQuery a few times a day however many people open it.

What `live.yml` needs is made by hand too, once, before it first runs on `main`. An environment a workflow names is otherwise made by its first run, with no rule on which branch may run in it.

1. **The App.** In the settings of the account or organisation that owns your copy, *Developer settings → GitHub Apps → New GitHub App*: any name and homepage, the webhook off, the repository permissions *Contents: Read and write* and *Pull requests: Read and write* and no others, installable on this account only. Note its client ID, and generate a private key.
2. **Install it** on your copy alone.
3. **The environments.** In the repository's *Settings → Environments*, make `live-numbers` and `live-numbers-delivery`, and limit each one's deployment branches to `main`. Give `live-numbers-delivery` the secret `LIVE_APP_PRIVATE_KEY`, the whole private key file, and the variable `LIVE_APP_CLIENT_ID`; then delete the key file you downloaded.
4. **Auto-merge.** *Settings → General → Pull Requests → Allow auto-merge*. If the rules of `main` ask for an approval, the App's pull request waits for one; with none asked, nothing but the required checks stands between the App and `main`, so keep its key in that one environment.
5. **`infra/ci.env`.** `GCP_LIVE_SERVICE_ACCOUNT` is the reader's address, which `terraform output live_service_account` prints once the delivery has made it.

## The load alert

Where `operator_email` names an address, the delivery creates an email channel to it and one alert: Cloud Run's own count of the service's requests, all its revisions together, averaged over the last five minutes, once it has stayed for a minute past `load_alert_per_minute` requests a minute, 300 unless set. A card waiting for a task asks fifteen times a minute, so 300 is some twenty families at once; the busiest minute of real lessons so far held 19 requests to the MCP endpoint on one instance. The mail names the commands that tell more, `just usage 1h` and `just report 1h`, and a second mail comes once the load is back under the line. The count is the platform's, so it costs nothing to keep and is there whatever the service's own telemetry does.

- **Before the first delivery with it**, grant the applying identity its two roles ([above](#when-the-bootstrap-changes)).
- **The address is written in the repository**, in the tfvars, so it has to be one already public, such as the one the privacy policy names. The mail comes from Google's alerting, and a group address has to accept mail from `alerting-noreply@google.com`.
- **What it costs**: nothing until Google starts charging for alerting, then some $0.39 a month ([above](#what-can-cost-money)).

## The daily report

Every morning Data Studio mails one page of the private report: what yesterday cost and how the service was used, beside the month so far and a table of the last fortnight. The costs come from Cloud Billing's export to BigQuery, the counts are the night's, from `impact.daily`, and one view, `impact_private.daily_report`, sets them side by side. It needs `analytics`, and is made by hand, once, in this order:

1. **Turn the export on.** In the Billing console of the account the project is billed to: *Billing export → BigQuery export → Standard usage cost → Edit settings*, the project, and the dataset `billing` the delivery has made. It takes *Billing Account Costs Manager* or *Administrator*. Turning it on lets the export's own identity write into the dataset; leave that right where it is. A dataset in a region is given the costs from the day the export is turned on, and none from before.
2. **Wait for its table.** The export writes its first rows within hours, sometimes a day, and `bq ls PROJECT_ID:billing` then lists `gcp_billing_export_v1_…`. Set `billing_export = true` in the tfvars and deliver it: the delivery makes the view. Set before the table is there, the delivery stops at the view, since BigQuery makes no view over a table that does not exist.
3. **The page.** In the private report, add data from BigQuery: the project, the dataset `impact_private`, the view `daily_report`. In the data source, set the data freshness to one hour, and the type of `charged`, `credits` and `cost` to the currency of the billing account. On a page of its own:
   - **Yesterday**: scorecards of `cost`, `charged`, `credits`, `learners`, `tasks`, `answers` and `topics_won`, each filtered on `yesterday` being true.
   - **The month so far**: scorecards of the sums of `cost`, `charged`, `tasks`, `answers` and `topics_won`, filtered on `this_month` being true, and of `learners_month` filtered on `yesterday`: a sum of the days' children would count a child once for every day.
   - **The last fortnight**: a table of `day` and the other columns, filtered on `days_ago` being at most 14, newest first.
4. **The schedule.** *Share → Schedule delivery*: the recipients, this page alone, a start at 8:00, repeated every day. Data Studio takes the time zone of the computer the schedule is made on and has no setting of its own, so make it on a computer set to the time zone the mail is wanted in.

What to know when reading it:

- A day is a UTC day, as every count here is, and so is a day of the costs: the one their usage began in, in UTC, rather than the Pacific day the Billing console shows. On the first of a month, the month so far is the month just ended, whole.
- The export promises no time of writing, and Google says a day's costs are usually there within a day: the morning's figure for yesterday may still grow, and the fortnight's table shows it settle.
- `charged` is the price of what was used, `credits` the free allowance and any other credit against it, below nought, and `cost` what is left to pay. They are this project's charges alone, as the spend alert counts them: what the billing account is charged for itself or for its other projects is not in them.
- An empty day is one the night did not count, or one the export holds nothing of; `just impact` lists the first.
- Moving the project to another billing account leaves the export behind: turn it on again on the new account, into the same dataset.

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

With `analytics` on, the destroy stops at the tables of the counts: they are years of evidence, and protected from deletion on purpose. To remove them too, set `deletion_protection = false` on them in `infra/terraform/analytics/main.tf`, apply, and destroy again. The log bucket goes, and its name stays taken for seven days, so `analytics` turned off and on again within a week fails until they pass. The destroy stops at the dataset `billing` too once the export has written into it, and so does a delivery that turns `analytics` off, since what it holds cannot be exported again: delete the export's table by hand, after turning the export off, to go on. The Data Studio reports and the schedule of the daily report were never described here either.
