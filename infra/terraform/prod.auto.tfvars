# The deployment this repository delivers to. Nothing here is secret: the client
# id travels in every authorization request, and the federated pool is guarded
# by a condition on the repository rather than by the secrecy of its name.
#
# Terraform loads this file by itself, because of the name. Changing where this
# repository deploys is a pull request, not a setting in somebody's console.

project_id = "mathtrail-prod"
region     = "us-central1"

# billing_account is deliberately not here. It names the account that pays, and
# this file is public; it arrives from the environment instead, as
# TF_VAR_billing_account.

# The host the service answers on. Its ownership is verified once, and the
# deployment identity is added as an owner beside it.
public_host = "mcp.mathtrail.app"

# From the Google OAuth client. Its secret is not here and never will be.
google_oauth_client_id = "312162576330-7vf0qup09aqltttpao4506rapp6ihnu3.apps.googleusercontent.com"

# Whose workflows may deploy, and the numeric id of that owner.
github_repository = "MathTrail/mathtrail-standalone"
github_owner_id   = "261630499"

# The pool created before the first apply, in which those workflows' tokens are
# exchanged for Google ones.
workload_identity_pool_id = "mathtrail-github"

# Turned on once the domain answers, so that the platform's own address for the
# service stops resolving and one entrance is left.
disable_default_url = true
