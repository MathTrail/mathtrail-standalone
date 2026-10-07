# What the service keeps, and where

As of 7 October 2026. Every place a family's data can be, who can read it there, for how long, how it is protected and how it is removed. This page is written for a school or a funder that asks; parents get the same facts in plainer words in the [privacy policy](https://mathtrail.app/en/privacy/). Keep it current when any of it changes.

## In brief

- The service is one program on Google Cloud Run. It keeps no database and no disk, and it remembers nothing between requests.
- The child is known by a pseudonym the adult chooses. No real name, birth date, school, address or photograph is asked for or kept anywhere.
- The child's profile is one file in the adult's own Google Drive. The service reaches it only while the adult's grant lasts, and only because the service made the file.
- The service's log records what each request did and, for each answer, whether it was right, the trap behind a wrong one, whether the hint was used and the child's chance of success, under the code of the adult's account. It never records a task, the letter chosen or the pseudonym. Its lines leave after 30 days, and a copy of three kinds of line after 62.
- What is kept for years are counts that hold no identifier of any kind.
- Over the lessons the chat's model sees nearly everything the profile holds, so all of it reaches the chat provider the family uses.

## Where each thing is

| What | Where | Who can read it | How long | How it is protected | How it is removed |
|---|---|---|---|---|---|
| The profile | `mathtrail-profile.json` in a `MathTrail` folder of the adult's Google Drive | The adult; the service while the adult's grant lasts; Google | Until the adult deletes it | The adult's Google account; the `drive.file` scope, which reaches only files the service made; the answer of the current task sealed under the service's key | The adult deletes the file and empties the Drive bin |
| A profile set aside | The same folder, under a new name, when a profile could not be read and the adult chose to start again | The same | Until the adult deletes it | The same | The same |
| Tokens | Wherever the chat host keeps them: an access token and a refresh token | Nobody without the service's key | 15 minutes for an access token; 30 days for a refresh token, renewed for at most 90 days from the sign-in | Sealed with XChaCha20-Poly1305 under keys kept in Secret Manager | They expire; the adult disconnects the app in the chat or removes its access in the Google account; a new key, with the old one dropped, ends all of them at once |
| Two sign-in cookies | The adult's browser, for `mcp.mathtrail.app` alone | The service | 10 minutes, taken back when the sign-in ends; 180 days for the list of apps the adult approved | Sealed; `HttpOnly`, `Secure`, `__Host-` | They expire, or the browser clears them |
| Counters of the pace of requests, kept by the address or by the account's code | The memory of the instance that answers | The service | While the instance runs; at most 4,096 keys for each limit, the longest unused forgotten first | Never written anywhere | The instance stops |
| The service's log | Cloud Logging, the `_Default` bucket, location `global` | The project's owners; the default Compute Engine identity | 30 days | Google Cloud's access control | Lines leave when their time is up; Cloud Logging deletes no single line |
| Cloud Run's request log | The same bucket | The same | 30 days | The same | The same |
| The copy of the lines children are counted from | Cloud Logging, the `activity` bucket in `us-central1`, which BigQuery reads as the dataset `activity_logs` | The project's owners; the nightly counting job; the identity the configuration is applied as; the default Compute Engine identity | 62 days | The same | The same |
| The counts | BigQuery, the dataset `impact` in `us-central1`, read through the views `impact_private` (exact numbers) and `impact_public` (numbers anyone may see); the latest month of the answers weighed against the chance promised, as the public views show it, is kept every day as one row in the dataset `impact_site`, copied once a month into the site's repository, `site/research/live.json`, and shown on the page "Research" | The project's owners, the counting job, the identity the configuration is applied as and the default Compute Engine identity; the public views anyone, once the public report is published; the row in `impact_site` also the identity `mathtrail-live`, which may read nothing else and which only the repository's monthly job that brings it to the site may borrow; and its monthly copy on the site anyone | For years | No identifier of any kind; the public views show closed months only and no group of fewer than ten children, and a range of the answers weighed against the chance promised only where ten children and thirty answers stand behind it; they round every count to the nearest five, and give shares, the chances promised and how far the answers came out from them to three decimal places, and other means to one | Nothing in them tells one child from another, so nothing is removed |
| Traces | Cloud Trace | The project's owners; the default Compute Engine identity | 30 days | The service traces about one request in ten, and blanks the address, the user agent, the path and the query before a trace leaves the process. Cloud Run also traces up to one request every ten seconds on each instance by itself, under Google's rules | They expire |
| Metrics | Cloud Monitoring | The project's owners; the default Compute Engine identity | 24 months | Labels from closed lists: the tool, the outcome, the route; nothing about a child | They expire |
| Keys | Secret Manager, replicated where Google chooses | The service reads them; the owners and the identity the configuration is applied as manage them | Until replaced | Google Cloud's access control; no key is in the repository, whose history is scanned for one on every change | A replaced version is disabled |
| The site's access log | GitHub Pages | GitHub | Under GitHub's policy | The site sets no cookies, runs no analytics and loads nothing from other domains | Under GitHub's policy |
| The conversation | The family's chat provider: Anthropic for Claude, OpenAI for ChatGPT | The provider | Under the provider's policy | Under the provider's policy | Under the provider's policy |

### What each holds

- **The profile:** the pseudonym, the grade, the child's interests, the skills to keep out of tasks and the adult's free-form notes, up to 500 characters; a random identifier of the child; the language of the lessons, if chosen; the topic the lessons are kept to, if chosen; the country and, in the United States, the state, if the adult gave them; the ratings, the topics mastered and a summary of past answers by topic, with the traps that came up; where the ratings stood before the latest answers and at the start of each of the last seven days with answers; a counter of today's tasks; a sketch of each of the last 200 tasks, so that new ones do not repeat them; the request for the next task while it is being written; the current task, its answer sealed; when the file was made and changed, and by which version.
- **A token:** an opaque code for the adult's account, made with the service's key from the Google account; Google's access token; and the country the adult signed in from, looked up at that moment in a database on the service's own server; the address itself is not kept in it. A refresh token adds Google's refresh token and when the adult signed in.
- **The service's log:** what each request did — the tool, the outcome, why a task was refused, the attempts, the durations, the size in bytes of each part of a task handed in, a limit reached — with the version of the instructions the model was given and the opaque code of the account. The line of an answer adds its topic, level and difficulty, whether it was right, the trap behind a wrong one, whether the hint was used, the chance of success the child's rating gave, who chose the task and the range the child's count of answers was in: while the log keeps them, these lines are the child's answers one by one, under the account's code. Three kinds of line, a task given, an answer recorded and a topic mastered, add between them a name the child is counted under for one calendar month, made with a key of its own from the child's identifier, which leads back neither to the child nor to the identifier; the chat, the language, the grade, the month the profile was made and the number of topics mastered; and the country, the state and the country of the sign-in, each as a short code. No line carries the pseudonym, a task, the letter chosen, an email address, a token, the child's identifier or the address a request came from.
- **Cloud Run's request log:** the standard record Google keeps of every request to any service on Cloud Run: the time, the address it came from, what the browser says it is, the page that sent it, the full address asked for, the status, and the size and duration of the answer. During the sign-in the address asked for carries what the chat and Google put in it: Google's one-time sign-in code, which only the service can redeem, the domain of a Google Workspace account, and the email address the chat suggests signing in with, if it suggests one.
- **The counts:** totals of each day, and of each week and month, told apart by one thing at a time — the chat, the language, the grade, the country, the state, the country of the sign-in, the month the profile was made — and, for each month, the traps that wrong answers fell into, by topic and grade, and the answers to the tasks the service chose weighed against the chance it promised, by range of that chance and by the range of the child's count of answers, with the sums a mean's error by child is read from. A total counts a group of children, and a group may be a single child: that is why no total under ten is ever shown publicly.

### Who can read in Google Cloud

- The project's two owner accounts, both the maintainer's.
- The service's own identity, which reads its three secrets and writes traces and metrics, and reads nothing back.
- The nightly counting job's identity, which reads the copy of the lines and writes the counts.
- The identity the delivery applies the configuration as. It manages everything above, so it can read the copy, the counts and the keys.
- The identity that deploys a new release, which reads no data.
- The default Compute Engine identity, which Google creates with the Editor role. Editor reaches the logs, the traces, the metrics and BigQuery, though not the keys. Nothing in the project runs as it, and taking the role away is still to be done.

## What the chat's model sees, tool by tool

Everything a tool returns is read by the chat's model, and so reaches the chat provider with the rest of the conversation.

| Tool | What its result shows the model |
|---|---|
| `get_profile`, `save_profile` | The child's details: the pseudonym, the grade, the interests, the skills kept out, the language of the lessons, the country and state if given, the topic the lessons are kept to; the adult's notes; how far the first five tasks have got; what comes next; the folder and the names of the profile's file and of any file set aside, with their links in Drive; the last answer recorded |
| `next_task` | Whether a request for a task is open, and since when; who the card is for: the pseudonym, the grade and the language chosen; the language of the lesson; the last answer |
| `get_package` | What to write the task from: the topic, the level, the difficulty and why the task comes now, where the child's chances lie in the topic, the traps to build the wrong options on, three reference tasks and the formats; the grade, the child's interests, the skills kept out and the adult's notes |
| `submit_task` | The task as the child's card shows it, with no answer, or every reason it was refused; the attempts left; who the card is for; the last answer |
| `submit_answer` | The answer given, whether it was right, the right option, the trap behind a wrong one, the solution, whether the hint was used and how the rating in the topic moved |
| `get_progress` | The ratings and how they moved since the last task and over the week, the topics mastered, the latest answers, the tasks left without an answer, the traps that keep coming back, what comes next, the child's details without the notes, and where the profile is |

The child's identifier is in no result, and the pseudonym never goes into the text of a task. The task the model hands in is its own writing, with the answer, the solution and the explanations of the wrong options, and the chat shows it in its log of tool calls: an adult who reads that log in front of the child shows the answer.

## Limits

- One address may send the sign-in's endpoints 20 requests a minute; a signed-in account may send the MCP endpoint 60 a minute; one instance takes 200 a minute from everybody for the sign-in, and 200 for the lessons; and one account's grant may be renewed at Google 12 times a minute. Each instance counts them in its own memory.
- A child may be given 20 tasks a day, and once 5 requests of a day have ended with the model out of attempts, no more are opened that day.

## When a key leaks

The service holds three secrets: the key tokens are sealed with, the key children are counted under, and the secret of the Google sign-in client. When one of them leaks, or may have, this is what is done.

- **The sealing key** opens nothing by itself. With a token a chat host holds, it opens the Google grant sealed inside, which reaches only the files the service made in that adult's Drive. A new version of the key is added and made the only one, with no previous key kept: the routine [rotation](../self-hosting.md#rotating-the-sealing-key), less the old key it keeps. Every token and cookie sealed with the leaked key stops working at once, every family signs in again, a task on a child's card at that moment is dropped, since its sealed answer can no longer be opened, and the leaked version is disabled.
- **The Google client secret** is what turns a Google refresh token taken out of a token into fresh access. A new secret is made at Google and given to the service, and the old one is deleted there; together with a new sealing key, that leaves a grant taken out good for an hour at most.
- **The key children are counted under** leads back to no child by itself: only together with a child's identifier and the log could it tell which lines are that child's. It is replaced at once, and the month in which that happens counts every child twice.
- **The notice.** A public notice goes on the site, in the README and as a GitHub security advisory: what leaked, when, what it could reach and what was done.

Any other exposure of a family's data is handled the same way: the cause is fixed and delivered first, what can be withdrawn is withdrawn, and the same notice goes out.

## Where the facts come from

- The buckets, datasets, secrets and identities: [`infra/terraform/`](../../infra/terraform/). The 30 days of the `_Default` bucket are Google's default, which this project's configuration does not set.
- The limits: [`internal/config/config.go`](../../internal/config/config.go).
- What each tool returns: [`internal/transport/mcp/`](../../internal/transport/mcp/); the package the model writes from: [`content/package.go`](../../content/package.go).
- What a trace leaves out: [`internal/telemetry/redact.go`](../../internal/telemetry/redact.go).
- The tokens, the cookies and the keys: SPEC section 9.3; the rotation: [self-hosting](../self-hosting.md#rotating-the-sealing-key).
- The retention of traces and metrics: Google's published quotas for Cloud Trace (30 days) and Cloud Monitoring (24 months for metrics sent by OpenTelemetry), read on 5 October 2026.
