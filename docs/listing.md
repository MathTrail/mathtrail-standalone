# What goes into the directories' forms

Everything the submission forms of Claude and ChatGPT ask for that is text, ready to paste. The requirements behind each field are rows of [catalogs.md](catalogs.md), named in brackets.

- **Claude** takes its listing in its portal, field by field: the texts are below.
- **ChatGPT** takes most of its listing from the plugin's package, [`plugin/plugin.json`](../plugin/plugin.json), when the ZIP is uploaded: the name, the subtitle, the description, the capabilities, the starter prompts, the five and three test cases, the release notes and the Russian translation. They are written there once and not repeated here. What the dashboard asks for apart from the package is below.

The texts describe the lesson as the author decided it on 2026-10-06 (О-61): the adult types in the chat, and the child, beside them, answers on the card. T84 makes the model's instructions match it, and the examples and cases are run live only after T84: Claude's in T82, ChatGPT's in T63 and T71.6.

Every length below is counted by the package's test: `go test ./plugin/ -run Limits -v` prints each count against its limit. The fenced blocks named for a field are what the test reads.

## Before submission

1. The developer is ALT EdTech (О-65), named on 2026-10-06: in the `company` block below, and in `plugin/plugin.json`, in `author.name` and in `developerName`. The package's test holds the three equal. ChatGPT shows the name OpenAI verifies, so if its verification names another, all three change together, and with them the privacy policy and the terms of item 2 (T71.6).
2. The privacy policy and the terms name ALT EdTech as the maker of MathTrail, in every language of the site. ChatGPT asks that the public pages "identify the same publisher as the submission" (GPT-05), and `just plugin-zip` refuses while they do not name the package's developer.
3. Refresh the demo profile and run the reviewer's scenario with the examples below (T82).
4. Build the ChatGPT package with `just plugin-zip`, and raise `version` in `plugin/plugin.json` for every new ZIP.

## Claude

### The listing (CL-04)

The name, at most 100 characters:

```text claude-name
MathTrail
```

The one-liner, at most 200 characters:

```text claude-one-liner
Olympiad-style maths for grades 1–6 in a lesson a parent or tutor runs: Claude writes each task, MathTrail checks it before the child answers on a card, and keeps the progress.
```

The description, at most 2,000 characters. Anthropic does not edit it; after publication it changes only by a listing edit, which a reviewer approves.

```text claude-description
MathTrail is for parents and tutors who want to give a child in grades 1 to 6 olympiad-style maths: tasks that ask for an idea rather than a drill.

How a lesson goes. You ask Claude for a task. Claude writes it, and MathTrail checks it with a program before your child sees it: exactly one of the five options is right, a solver that tries every case reproduces the answer, the wording suits the grade, the task is not a near copy of an earlier one, and any picture is well formed, labelled as the question names it and has no answer written on it. The task then appears on a card, and your child, sitting beside you, answers by tapping an option. Whether the wording, the picture and the solution agree in meaning is checked by Claude itself.

After an answer. The card asks Claude to go over the answer, and how it went comes below, on a card of its own: the mistake behind a wrong answer, a picture of the solution where the task has one, and the solution step by step. The first five tasks find where your child stands; after them, MathTrail chooses each task's topic and difficulty from the answers. Ask how your child is doing to see the rank in each topic, what moved over the past week, the mistakes that repeat and what to practise next.

Your child's data. You sign in with Google, and your child needs no account. The profile is one file in your own Google Drive, under a pseudonym, never a real name, and MathTrail reaches only the files it created there.

Limits. Up to 20 new tasks a day. MathTrail is not a homework solver, and it covers grades 1 to 6 only. Its code is open source under the MIT License.
```

The categories, one to five, from the portal's own list, which its documentation does not publish (T87 picks from it):

```text claude-categories
Education
```

- **The slug**, locked once published, chosen by the author on 2026-10-06: `mathtrail`.
- **Documentation:** https://mathtrail.app/en/help/ (CL-13, R229).
- **Privacy policy:** https://mathtrail.app/en/privacy/ (CL-20).
- **Support contact:** altedtech.info@gmail.com (CL-21).
- **The icon:** `plugin/assets/logo.png`, 1024 pixels square. The portal's rule for it is not documented; a public report quotes "Square PNG or JPEG, 512 to 2048 px per side, under 2 MB", which it meets (§10 of catalogs.md).

### Use cases (CL-05)

- **What it is for:** a lesson of olympiad-style maths for a child in grades 1 to 6, run by a parent or a tutor in their own chat; the child's progress and what to practise next; the child's profile, and a topic to keep the lessons to.
- **What users need before they connect:** a Claude account and a Google account. The child needs neither.
- **Whether it reads or writes:** both. It reads and writes one file, the child's profile, which it creates in the signed-in adult's Google Drive; the `drive.file` permission reaches no other file.

### Company (CL-06)

```text company
ALT EdTech
```

- **Website:** https://mathtrail.app
- **Contact for review updates:** altedtech.info@gmail.com

### Authentication (CL-07)

OAuth 2.1 with PKCE, by a client ID metadata document or by dynamic client registration, at `https://mcp.mathtrail.app`. The adult then signs in with Google, which is asked for `drive.file` alone (R225).

### Data handling (CL-08)

- **The API is MathTrail's own:** one service on Google Cloud Run, with no database. The child's profile is stored in the parent's own Google Drive, on the parent's grant.
- **No personal health data**, and **no sponsored content** in the connector. The site's menu leads to a page about the author's coming product, with its price; the author accepted that as it is (О-64).

### Test and launch (CL-09)

- **The account:** the reviewers' sign-in of [reviewers.md](reviewers.md). On MathTrail's consent screen, the reviewer opens "Reviewing MathTrail for a directory?" and types the password, with no Google step. It opens a demo profile with a filled progress.
- **The password** is typed into the form from the password manager, and is written nowhere else.
- **The steps to try:** the three examples below, and "What to try" in [reviewers.md](reviewers.md).

### Allowed link URIs (CL-12)

`https://mathtrail.app`. The card links only to the site's pages (R186), and nothing else of the project's is a link target.

### Three working examples (CL-24)

1. **"Let's do a MathTrail task."** A card appears at once and shows that the task is being prepared, while Claude writes the task and MathTrail checks it (`next_task`, `get_package`, `submit_task`). Within a minute or two the task arrives on the card with five options. The child taps one, and the card asks Claude to go over the answer: how it went comes below, on a card of its own, with the mistake behind a wrong answer, a picture of the solution and the solution step by step (`submit_answer`, `show_result`). Meanwhile Claude writes the next task ahead (`prepare_task`, `submit_task`), so "Another task", sent to the chat, brings it at once, on a new card below (`next_task`).
2. **"How is my child doing?"** A card shows the rank in each topic, what moved since the last task and over the past week, a review of strengths and topics to develop, and the mistakes that repeat; Claude sums it up for the adult (`get_progress`).
3. **"Change my child's interests to space and dinosaurs."** Claude asks to allow MathTrail to save the profile, since saving replaces what was kept, then says what was saved (`save_profile`). The next tasks may draw on the new interests.

### Screenshots (CL-11)

PNG, at least 1000 pixels wide, three to five, of the card alone and without the prompt in the picture. `just listing-shots` takes them from the widget's preview into `web/listing/`, 1280 pixels wide, with the card's version left out and the task's row of posts cut short, in the order and the themes below (R232). Each picture's prompt is given beside it. No picture shows an answer before the child has answered, or the chat's log of tool calls (О-27).

1. "Let's do a MathTrail task." The card with the task and its five options, in the light theme.
2. "Let's do a MathTrail task.", then the bulb, Hint, tapped on the card. The same card with the hint open, in the dark theme.
3. "Let's do a MathTrail task.", then a wrong option tapped on the card, and "Go over the answer" sent. The card of how the answer went names the mistake and shows the picture of the solution and the solution, in the light theme.
4. "How is my child doing?" The progress with its review open, in the dark theme.

## ChatGPT

### From the package

`plugin/plugin.json` and `plugin/mcp.json` are the package (GPT-03, GPT-04, GPT-10, GPT-14). `just plugin-zip` checks them and builds the ZIP into `bin/`. The package's own `version` goes up with every new ZIP; the service's releases need none (GPT-50).

### The hints of the tools (GPT-12)

ChatGPT asks for a justification of every hint in its dashboard, after it scans the tools. They are written from R215, R214, R100, R101 and R246.

- **Every tool, `openWorldHint: false`:** the tool reaches only MathTrail's own service and the one file MathTrail created in the signed-in adult's Google Drive, under the `drive.file` permission. Nothing reaches the public internet or anyone outside the account, and nothing is published.
- **Every tool that reads (`get_profile`, `get_progress`, `get_package`, `show_result`; and `read_progress` and `read_task`, which only the card calls), `readOnlyHint: true` and `destructiveHint: false`:** the tool reads the profile's file, and the profile's tool and the two of the progress also look up the file's folder to say where it is. It changes nothing, a damaged file included: it reports a damaged file and leaves it as it is. Each call leaves one line in MathTrail's own operational log, as every request to the service does. That line is the service's record, not a change to the user's data.
- **`next_task`, `readOnlyHint: false`, `destructiveHint: false`, `idempotentHint: false`:** it opens a request for a task in the profile's file, and a task left on the card unanswered is recorded as skipped, the history growing by an entry. It never replaces what the adult set and never deletes a state of the file. Called again while a request is less than fifteen minutes old, it hands back the same request and writes nothing; called again once that task is on the card, it records the task as skipped and opens another request, so a second call is not the same as one.
- **`prepare_task`, `readOnlyHint: false`, `destructiveHint: false`, `idempotentHint: true`:** it opens a request for the next task, written ahead, in the profile's file, and hands the model what to write it from; it lets go of a task written ahead that no longer fits the lesson, which the child never saw. It never replaces what the adult set and never deletes a state of the file. Called again while the request is open, it hands back the same request and writes nothing.
- **`submit_task`, `readOnlyHint: false`, `destructiveHint: false`, `idempotentHint: false`:** it records the attempt it spends, one of three, and once the task is accepted, the task with its answer sealed — on the card, or kept for later when it was written ahead. It replaces and deletes nothing. It is not idempotent because each call spends an attempt, which its description says.
- **`submit_answer`, `readOnlyHint: false`, `destructiveHint: false`, `idempotentHint: true`:** it adds the answer to the profile's record, with the ratings and counts it moves. Nothing the adult set is replaced. The answer is recorded once: a repeat writes nothing and returns the first result. When a task's sealed answer can no longer be opened, it takes the task off the card and records that a new one is owed; nothing set is lost.
- **`save_profile`, `readOnlyHint: false`, `destructiveHint: true`, `idempotentHint: true`:** each field given replaces the value saved before, which is then gone from the profile; `start_over` and `restore` replace the whole profile and keep the old file in the adult's Drive. The safeguards: the model calls it only when the adult asks, the host asks before a destructive tool runs, and Drive's version history keeps the first version of each day the profile changed.
- **`edit_profile`, `readOnlyHint: false`, `destructiveHint: true`, `idempotentHint: true`:** only the card calls it, when the adult saves the profile's form or chooses a topic on the card; the fields sent replace the values saved before.

### Reviewer access (GPT-09)

The same sign-in as Claude's, from [reviewers.md](reviewers.md): the reviewer opens "Reviewing MathTrail for a directory?" on MathTrail's consent screen and types the password, which goes into the dashboard's secure form and nowhere else. The package carries no credentials and no reviewer instructions: ChatGPT refuses `test_credentials` and `reviewer_instructions` in a ZIP.

### Left to T71.6

- the demo recording's address (`review.demo_recording_url`, GPT-10);
- the countries (`publication.countries`, GPT-14);
- the domain's token, served from a setting (R216, GPT-08);
- the developer's verified identity (О-65, GPT-01).
