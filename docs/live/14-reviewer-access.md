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
