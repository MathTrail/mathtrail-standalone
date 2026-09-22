# Security policy

MathTrail is an olympiad maths coach for children in grades 1–6, run inside somebody's own Claude or ChatGPT. It is built around a child, so a fault here is not only a technical one, and a report about it is welcome from anybody who finds one.

## Reporting a vulnerability

Report it privately, through GitHub:

**[Report a vulnerability](https://github.com/MathTrail/mathtrail-standalone/security/advisories/new)** — or open the repository's **Security** tab and choose *Report a vulnerability*.

The thread is private until it is resolved, it stays attached to the repository, and an advisory with a CVE can be published from it when the finding warrants one. Please do not open a public issue for something exploitable.

There is no other reporting channel: this project is run by one person, and a private thread here is the only place that will be read.

What helps, in whatever detail you have it: what you did, what happened, what you expected instead, and how an attacker would use it. A proof of concept is welcome and never required.

**What to expect.** This is a free project maintained in spare time. There is no response-time guarantee. A report that lands will get an acknowledgement, a fix for what is exploitable, and credit in the advisory unless you ask otherwise.

## What is in scope

- The service at `mcp.mathtrail.app` — the MCP endpoint, the sign-in and the tokens it issues.
- The site at `mathtrail.app`.
- Everything in this repository, including the workflows that build and deploy it.

Out of scope: Google's own services, the chat hosts (Claude, ChatGPT) and anything running on a fork.

## Testing, and what not to do

Research in good faith is welcome, and nobody will be pursued for it. Two limits, and both are about people rather than about us:

- **Only your own accounts and your own data.** No access to anybody else's Google account, Drive file or child's profile — a profile is a real child's.
- **No load testing, no denial of service, no spam of the sign-in.** The service runs inside a free tier: flooding it takes the lesson away from whoever is in the middle of one.

## What the service holds, and what it does not

Worth knowing before you look for something that is not there:

- **Nothing is stored between requests.** There is no database and no disk. The child's profile is a JSON file in the *parent's* own Google Drive, reached with the `drive.file` scope, which grants access only to files this application created.
- **Only a pseudonym.** No real name, no birth date, no school. None of it is asked for, and the pseudonym never reaches the text of a task.
- **No personal data in the logs.** Log lines carry aggregates — which tool, which outcome, how long, how many attempts — never a task, an answer or a profile.
- **The answer to the current task stays sealed** until the child has answered: it is encrypted with the service's own key and is not in any widget payload, in any open field of the profile file, or in any log.
- **No secrets in this repository.** Keys live in Secret Manager and reach the service through its environment. The whole history is scanned on every pull request, and a key that was ever committed would be treated as leaked and rotated.

## Supported versions

The tip of `main` is what is deployed and what is supported. There is no separate maintenance branch: a fix goes to `main` and is delivered from there.

## How this is kept honest

Every change arrives through a pull request that cannot be merged until the following pass: the tests with the race detector, the linter including `gosec`, CodeQL, `govulncheck`, a secret scan of the whole history, a licence check, a vulnerability scan of the runtime image, and 90% coverage of the lines the change touches. The reasoning behind each of these is written down in [`docs/decisions.md`](docs/decisions.md).
