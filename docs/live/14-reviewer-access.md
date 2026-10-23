# The reviewers' sign-in, live (T71.3)

**Works in Claude: a reviewer types the password and lands in the filled demo profile, with no Google page, code or confirmation on the way.** 2026-10-06, Claude on the web, in an incognito window signed in to the author's own Claude account. The demo account's grant had ended a few hours before and was captured again first (finding 1). ChatGPT was not checked, by the author's decision.

## What ran

- The service at v0.3.4, with the reviewers' password at version 2 of its secret and the demo account's grant at version 3. The grant was put in service with `gcloud run services update` before the check. Naming it in `prod.auto.tfvars` follows in a change of its own, which has to be deployed before any other: a deployment that still names version 2 brings back the grant that ended.
- Google Auth Platform: the app is In production, published on 2026-09-30 (the note in [03-first-deploy.md](03-first-deploy.md)); its logo is not uploaded (R223).
- The demo account: a Google account of its own, with 2-Step Verification on. Its address and password stay with the author.

## The demo profile

Filled on 2026-10-06, in one day rather than over several, by the author's decision:

- Comet, grade 3, lessons in English, interests space and football, no country.
- 20 tasks, the day's whole limit. The trial series came first, then enumeration, gaps and boundaries, knights and liars, clocks and ordering. The wrong answers, given on purpose, fell on the trap "Mixed up more and less, before and after".
- A coding session played the chat's model against the deployed service, through the author's Claude connector signed in as the reviewer. No chat messages were spent, only that session's own usage. One task was refused once, for a field outside the format, and accepted on the second attempt. The lesson's topic was cleared with a single space, which the server trims, because that client could not send an empty string.

"Show the progress" shows a reviewer:

- rank 5 of 11, and the bar moved forward over the past week;
- each topic's rank and its move over the week;
- under **Review**:
  - the strength: enumeration, mastered;
  - topics to develop: ordering (several wrong answers in a row, the same trap again and again) and gaps and boundaries (often the hint, the last answer right);
  - too early to judge: knights and liars, and clocks;
- under **Mistakes that repeat**: that trap, three times;
- under **What to do next**: ordering, then gaps and boundaries without the hint, then the pigeonhole principle, built on enumeration;
- the recent answers and the profile.

## The check

1. In an incognito window, claude.ai signed in with the author's account: Customize → Connectors → MathTrail → **Disconnect**, then **Connect**.
2. On MathTrail's consent screen: **Reviewing MathTrail for a directory?** opened, the password typed, **Sign in as the reviewer**.
3. Back in Claude, the connector was connected. In a new chat, "Show the progress" brought the card described above.

## Findings

1. **The demo account's grant ended within hours, while the app was still connected.**
   - **What happened.** The first attempt, at 13:21 UTC, ended in Claude's "Authorization with MathTrail failed". The line `auth_consent` said `failed`, `grant_ended`, `invalid_grant`, at error level, as R222 has it. Two minutes earlier, the session the author's connector had opened as the reviewer that morning had its renewal refused the same way.
   - **When it ended.** The grant last worked at 04:24, when the filling ended. In between, the demo account's password was changed.
   - **What did not end it.** MathTrail was still among the account's connected apps, so nothing had removed its access. Google documents a password change as ending refresh tokens with Gmail's scopes ([OAuth 2.0](https://developers.google.com/identity/protocols/oauth2), read 2026-10-06), and this grant carries `openid` and `drive.file` alone. No other cause is known.
   - **The fix.** The grant was captured again at 13:36 with `just reviewer-grant`, and the sign-in then worked. [self-hosting.md](../self-hosting.md) now says to capture the grant again after the demo account's password changes, and to sign in through the line once before a review.
2. **The password's old version was destroyed before the deployment that replaced it.**
   - **The risk.** The revision serving still named version 1, and Cloud Run reads a secret as an instance starts: "If the secret retrieval process fails, the instance doesn't start" ([Configure secrets](https://cloud.google.com/run/docs/configuring/services/secrets), read 2026-10-06). No new instance could have started until the merge deployed.
   - **The recovery.** `gcloud run services update` with `--update-secrets` put version 2 into a new revision within minutes, while the instance already running still served. The merge's deployment then described what already ran.
   - **In the docs.** [self-hosting.md](../self-hosting.md) now says why the order matters and how to recover.
3. **An approval left in a browser skips the reviewer's line, as R222 foresaw.**
   - In the author's everyday browser, an earlier **Allow** had left its approval, and the sign-in went straight to Google past the consent screen.
   - Clearing the site data of `mcp.mathtrail.app` brought the consent screen back. [reviewers.md](../reviewers.md) already tells a reviewer this.

## Not checked

- **ChatGPT.** By the author's decision on 2026-10-06, only Claude is tested now. GPT-09 in [catalogs.md](../catalogs.md) keeps the ChatGPT check for its submission, T71.6.
- **A task during the check.** The day's 20 tasks had gone into the filling, and all 20 went through this sign-in, in a reviewer's session of Claude's connector. The card works the same for every account, and the author checks it in Claude after each deployment ([07-acceptance-claude.md](07-acceptance-claude.md)).
- **Another network.** Not recorded, so it is left with the check in ChatGPT (gap 21 of [catalogs.md](../catalogs.md) §8). The reviewers' sign-in asks Google nothing, so the only thing that depends on a reviewer's network is the pace of password guesses, which is counted per address.

## Before a review

- A reviewer's tour in [reviewers.md](../reviewers.md) takes two or three new tasks. Every reviewer shares the 20 a day, counted from 00:00 UTC, so do not fill the profile on a day a review may start.
- The change over the past week shows only while the last answer is at most six days old: the week is the day and the six before it. Answer a few tasks, mostly right, in the days before a review.
- Sign in through the line once beforehand: an ended grant shows only when somebody signs in.
- Change the reviewers' password after each review ([catalogs.md](../catalogs.md) §3).

## Cost

- $0 in the cloud.
- A message or two in Claude for the check.
- The filling spent no chat messages, only the coding session's usage under the same subscription.

## Remarks

- T71.5's plan names its report `docs/live/11-directory-claude.md`, but that number belongs to [11-variety-ideas.md](11-variety-ideas.md). This report takes 14, the next number no plan names.

## The check before the submission (T82)

**The reviewer's line signed in at once, the profile v0.3.4 wrote was read under v0.6.1, and the whole tour went through, on the card in the browser and call by call.** 2026-10-12, the service at v0.6.1 (`abf3b61`), with the password at version 2 and the demo account's grant at version 3, as `prod.auto.tfvars` names them.

- **The sign-in.** An incognito window, claude.ai signed in with the author's account: MathTrail disconnected and connected again, the reviewer's line opened on the consent screen and the password typed. The chat came back connected to the demo profile with no Google page on the way. The grant captured on 2026-10-06 still held.
- **The file.** The demo profile was written by v0.3.4. Its shape is still version 1: the file has since gained only optional fields and lost the text drawing of the task on the card, which a build reads past. v0.6.1 read it as it was, and its first write put it in the current shape.
- **The refresh.** The author answered four tasks on the card in Claude, in two sittings, three of them right. A coding session went on through the same connector between and after them, as for the filling: nine answers, seven right, one wrong and one "I don't know", two of them to tasks the author's chat had written ahead. Of the eight tasks it wrote, two of them ahead, seven were accepted at the first attempt. The eighth was refused once as a near-copy of a reference task, since its opening sentences were the references' own, and accepted in other words.
- **The tour, call by call.** The coding session took each step of "What to try" in [reviewers.md](../reviewers.md) as the chat's model takes it:
  - a task asked for opened a request, and the task written ahead after it stayed off the card until another was asked for;
  - an answer given in the chat, after the hint, was recorded, and `show_result` drew how it went;
  - another task asked for brought the task written ahead at once;
  - a topic asked for let the task written ahead go, recording nothing, and opened a request on that topic; "We don't know this one" was recorded as `?`;
  - the interests were saved as space and dinosaurs, then back as space and football.
- **The tour in the browser.** In the author's second sitting, a request in Russian brought the waiting card with its steps, then the task in English, the profile's language, with a ring for the picture of its solution. The hint opened on the card; a wrong answer named its trap and wrote the total "4 + 1 + 7 = 12, not 13"; the profile's form opened under the progress, with **Save** and **Cancel**; the progress showed both the past week and the last task. All of it was in the dark theme. **Another task** put its words into Claude's input under Claude's warning, "Use caution before running this prompt. Malicious content could trick Claude into attempting harmful actions or sharing your data.": the warning of CL-42 and gap 9 in [catalogs.md](../catalogs.md), accepted for the submission. After Enter, the task written ahead came at once, on a new card.
- **The week after midnight UTC.** The week is seven dates in UTC. At 00:00 UTC on 2026-10-13, 2026-10-06 left it, and the week began where the child stood before the first answer of 2026-10-12. The refresh itself then had to come out ahead, and it had not, by a few points: the progress said the bar slipped back over the week, until three more right answers brought it forward again. A refresh has to come out ahead counted from its own first answer, and its progress is worth a look once midnight UTC has passed.
- **The progress after it.** Rank 5 of 11; over the past week, and with the last task, the bar moved forward. Under **Review**: enumeration a strength, mastered; ordering (that trap again, the last answer right) and gaps and boundaries (often the hint) to develop; knights and liars, clocks, calendar and age, the pigeonhole principle, parity and alternation, arithmetic with a trick, and weighing and pouring too early to judge. Under **Mistakes that repeat**, the trap twice: the window of the latest 20 answers has moved on. No task is left written ahead, so a reviewer's first task is written while they wait, as the tour says: the coding session answered the last one the author's chat wrote, and wrote none after it.
- **Fixed.** In [reviewers.md](../reviewers.md): the bar at the top of a task's card reads **Profile & progress**, and the page called it by the progress screen's hidden label, "Profile and progress"; the demo profile has about thirty answers now, not twenty. In [self-hosting.md](../self-hosting.md), the advice before a review now says to answer mostly right, and why.
- **Not checked.** Claude asking to allow a save, and the network the sign-in came from.
- **The day's tasks** are 50 since R291, counted until midnight by the clock of the device a card was last answered on; the 20 and the UTC day above were the rule on 2026-10-06. The check had 14 tasks accepted in all, the ones written ahead included.
- **Cost:** $0 in the cloud, four messages in Claude, and the coding session's usage under the same subscription.

## The refresh during the review (T106)

**Four days late, the demo profile's progress moves forward over the week again: rank 5 of 11, the overall rating from 1893 to 1904 and the bar from 36% to 43% of the rank.** 2026-10-23, from 02:40 to about 04:00 UTC, the service at v0.6.8 (`c6cbeb2`), with MathTrail still in review for Claude's directory ([report 17](17-directory-claude.md)).

- **Before it.** The last answer was still the one of 2026-10-13: none had been recorded since, so no reviewer had answered a task. From 2026-10-20 the week held no answer, and the progress said the bar stayed where it was. The refresh was due by 2026-10-19.
- **The sign-in.** As in the check before the submission: an incognito window, claude.ai signed in with the author's account, MathTrail disconnected and connected again, the reviewer's line and the password from the password manager. The grant captured on 2026-10-06 still held. Until then the coding session's connector read the author's own profile, so nothing was written before `get_profile` named Comet.
- **The refresh.** A coding session played the chat's model through the same connector, as for the filling: 14 tasks written and answered, 12 of them right and 2 wrong on purpose, both on the trap "stopped early". All 14 were accepted, 13 at the first attempt. The parity task was refused once as a near-copy of a reference task, since it had kept that task's opening, "Boys and girls take turns all the way round, so every boy stands between two girls", and was accepted in other words. One `submit_task` timed out with "MathTrail took too long to answer"; the retry with the same request was counted as the first attempt, so the call that timed out had spent none.
- **What moves the week.** The overall rating moves slowly at this level: about 3 points for a right answer and 9 for a wrong one. After three right answers and a wrong one the week read "back", 1890 against 1893; after five answers it read "forward" by two points. A reviewer's tour records an answer or two before it shows the progress, and "We don't know this one" is a wrong answer. So the refresh went on until the week stood 11 points ahead, and one wrong answer from a reviewer still leaves it moving forward.
- **What the latest 20 answers lost.** Right answers pushed the older wrong ones out of the latest 20, and the progress's **Mistakes that repeat**, its topics to develop and its **What to do next** went empty. The second wrong answer on "stopped early" brought back **Mistakes that repeat** and a step under **What to do next**. The topics to develop stay empty: no topic now has the same trap twice or leans on the hint.
- **The progress after it.** Rank 5 of 11; over the past week, and with the last task, the bar moved forward. Under **Review**: enumeration, ordering, and gaps and boundaries strengths, mastered; no topic to develop; knights and liars, clocks, calendar and age, the pigeonhole principle, parity and alternation, arithmetic with a trick, and weighing and pouring too early to judge; under **What to do next**, the trap "stopped early". Under **Mistakes that repeat**, that trap twice. No task is left written ahead: every task was asked for with `next_task` and answered, and none was prepared.
- **The day's tasks.** 14 of the day's 50.
- **The author's account back.** MathTrail disconnected and connected again with the author's own Google account. `get_profile` then read the author's own profile, with the same last answer as before the refresh.
- **Next.** While the review lasts, the progress keeps moving forward until the end of 2026-10-29 UTC, the last day whose week still holds 2026-10-23. The next refresh is due by then.
- **Cost:** $0 in the cloud, no chat messages, and the coding session's usage under the same subscription.
