#!/usr/bin/env bash
# Creates the few things that cannot be created by the pipeline that needs them:
# the project, the bucket its state lives in, and the federated identity the
# pipeline signs in as. Everything else about the deployment is Terraform, and
# Terraform runs in the pipeline.
#
# Safe to run again: every step checks before it creates.
#
# It reads the deployment's facts from infra/terraform/, so there is one place
# where they are written down, and prints at the end the two lines that a
# workflow needs before it can sign in at all.

set -euo pipefail

here="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
tfvars="$here/terraform/prod.auto.tfvars"
backend="$here/terraform/backend.hcl"

# One value out of an HCL file: the first `name = "value"` it finds.
hcl_value() {
    local name="$1" file="$2"
    sed -n "s/^[[:space:]]*${name}[[:space:]]*=[[:space:]]*\"\([^\"]*\)\".*/\1/p" "$file" | head -1
}

require() {
    local name="$1" value="$2"
    if [[ -z "$value" || "$value" == REPLACE* ]]; then
        echo "bootstrap: ${name} is not filled in yet (see ${tfvars##*/} and ${backend##*/})" >&2
        exit 1
    fi
}

project="$(hcl_value project_id "$tfvars")"
region="$(hcl_value region "$tfvars")"
repository="$(hcl_value github_repository "$tfvars")"
owner_id="$(hcl_value github_owner_id "$tfvars")"
pool="$(hcl_value workload_identity_pool_id "$tfvars")"
bucket="$(hcl_value bucket "$backend")"

require project_id "$project"
require region "$region"
require github_repository "$repository"
require github_owner_id "$owner_id"
require workload_identity_pool_id "$pool"
require bucket "$bucket"

# The account that pays is not written down beside the rest of the deployment,
# because that is written down in public. It arrives here the same way it
# arrives in a delivery.
billing="${TF_VAR_billing_account:-}"
if [[ -z "$billing" ]]; then
    echo "bootstrap: TF_VAR_billing_account is not set." >&2
    echo "It is the id of the billing account to link, of the form 01ABCD-234567-89EFGH." >&2
    exit 1
fi

service_account="mathtrail-tf@${project}.iam.gserviceaccount.com"

echo "==> project ${project}"
if ! gcloud projects describe "$project" > /dev/null 2>&1; then
    gcloud projects create "$project" --name=MathTrail
fi
gcloud billing projects link "$project" --billing-account="$billing" > /dev/null

# Without these nothing below can be created, and they are the reason this script
# exists rather than one more Terraform file.
echo "==> the APIs this script itself needs"
gcloud services enable \
    serviceusage.googleapis.com \
    cloudresourcemanager.googleapis.com \
    iam.googleapis.com \
    sts.googleapis.com \
    iamcredentials.googleapis.com \
    --project="$project"

echo "==> state bucket gs://${bucket}"
if ! gcloud storage buckets describe "gs://${bucket}" > /dev/null 2>&1; then
    gcloud storage buckets create "gs://${bucket}" \
        --project="$project" \
        --location="$region" \
        --uniform-bucket-level-access \
        --public-access-prevention
fi
# Versioning is what lets a state file that was written badly be rolled back.
gcloud storage buckets update "gs://${bucket}" --versioning > /dev/null

echo "==> federated pool ${pool}"
if ! gcloud iam workload-identity-pools describe "$pool" \
    --project="$project" --location=global > /dev/null 2>&1; then
    gcloud iam workload-identity-pools create "$pool" \
        --project="$project" \
        --location=global \
        --display-name="GitHub Actions"
fi

echo "==> pool provider github"
# Both halves of the condition matter. The name says which repository, and the
# numeric owner id says it is still ours: an account that is renamed or deleted
# frees its name for anyone to take, and the number goes with it.
if ! gcloud iam workload-identity-pools providers describe github \
    --project="$project" --location=global --workload-identity-pool="$pool" > /dev/null 2>&1; then
    gcloud iam workload-identity-pools providers create-oidc github \
        --project="$project" \
        --location=global \
        --workload-identity-pool="$pool" \
        --display-name="GitHub Actions" \
        --issuer-uri="https://token.actions.githubusercontent.com" \
        --attribute-mapping="google.subject=assertion.sub,attribute.repository=assertion.repository" \
        --attribute-condition="assertion.repository=='${repository}' && assertion.repository_owner_id=='${owner_id}'"
fi

echo "==> identity ${service_account}"
if ! gcloud iam service-accounts describe "$service_account" --project="$project" > /dev/null 2>&1; then
    gcloud iam service-accounts create mathtrail-tf \
        --project="$project" \
        --display-name="Identity the infrastructure is applied as"
fi

# Enumerated rather than Owner: enough to create everything the configuration
# describes, and nothing beyond this one project. The pool above is deliberately
# not among them — what this identity signs in through is not its to change.
# Roles may be written, and granted on the project itself, because the
# configuration defines both: the deployment identity gets a role of five
# permissions instead of a ready-made one of sixty, and the runtime identity
# gets the two that let it send telemetry about itself.
echo "==> what it may do"
for role in \
    roles/serviceusage.serviceUsageAdmin \
    roles/artifactregistry.admin \
    roles/secretmanager.admin \
    roles/run.admin \
    roles/iam.serviceAccountAdmin \
    roles/iam.serviceAccountUser \
    roles/iam.roleAdmin \
    roles/resourcemanager.projectIamAdmin \
    roles/browser
do
    gcloud projects add-iam-policy-binding "$project" \
        --member="serviceAccount:${service_account}" \
        --role="$role" \
        --condition=None \
        --quiet > /dev/null
    echo "    $role"
done

# The state, and the spend alert, sit outside the project's own policy.
gcloud storage buckets add-iam-policy-binding "gs://${bucket}" \
    --member="serviceAccount:${service_account}" \
    --role=roles/storage.objectAdmin \
    --quiet > /dev/null
echo "    roles/storage.objectAdmin on gs://${bucket}"

gcloud billing accounts add-iam-policy-binding "$billing" \
    --member="serviceAccount:${service_account}" \
    --role=roles/billing.costsManager \
    --quiet > /dev/null
echo "    roles/billing.costsManager on ${billing}"

echo "==> who may borrow it"
project_number="$(gcloud projects describe "$project" --format='value(projectNumber)')"
principal="principalSet://iam.googleapis.com/projects/${project_number}/locations/global/workloadIdentityPools/${pool}/attribute.repository/${repository}"
gcloud iam service-accounts add-iam-policy-binding "$service_account" \
    --project="$project" \
    --member="$principal" \
    --role=roles/iam.workloadIdentityUser \
    --quiet > /dev/null

provider="projects/${project_number}/locations/global/workloadIdentityPools/${pool}/providers/github"

cat <<SUMMARY

Done. Put these two lines in infra/ci.env and commit them:

GCP_WORKLOAD_IDENTITY_PROVIDER=${provider}
GCP_TERRAFORM_SERVICE_ACCOUNT=${service_account}

Three things are left that no API can do, and each is done once:

  1. The consent screen and a Web OAuth client, in the Google Auth Platform
     console. Its redirect URI is https://<your host>/oauth/callback; its client
     id goes into infra/terraform/prod.auto.tfvars.
  2. Two DNS records at the registrar: the TXT record Search Console asks for,
     and a CNAME from the service's host to ghs.googlehosted.com. Then, in
     Search Console, add ${service_account} as a verified owner of the domain:
     it is the identity that creates the domain mapping, and Cloud Run maps a
     domain only for an owner of it.
  3. The OAuth client's secret, as the repository secret
     GOOGLE_OAUTH_CLIENT_SECRET. It is the one value that exists nowhere else.

After that: merge to the main branch, and the deployment happens by itself.
SUMMARY
