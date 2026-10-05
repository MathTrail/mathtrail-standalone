---
title: Privacy policy
description: What MathTrail knows about your child, where it is kept, and how to take it back or delete it.
---

# Privacy policy

Last updated: 5 October 2026.

MathTrail is a free, open-source project. Questions about this policy or about your child's data go to [altedtech.info@gmail.com](mailto:altedtech.info@gmail.com).

## The short version

- **We store nothing.** The service keeps no database and no files. Between one request and the next it remembers nothing.
- **Your child's profile lives in your own Google Drive**, in a file you can open, copy or delete at any moment.
- **A pseudonym, never a real name.** No birth date, no school, no photograph.
- **We count children without knowing who they are.** A line of the log names a child by a code that changes every month and leads back to no one; what is kept for longer than two months are totals.
- **Nothing is sold, and nothing trains a model.**
- **Deleting the file deletes the profile.** There is no second copy anywhere.

## What MathTrail is

MathTrail is an app you add to your own Claude or ChatGPT. It picks a topic and a difficulty for your child, asks the chat's own model to write a task, checks that task with a program, shows it, records the answer and explains the mistake.

MathTrail runs no language model of its own. The model that writes the tasks is the one inside the chat you are already using, and your relationship with that chat's provider is governed by their terms and their privacy policy, not by this one.

## Who signs in

Only an adult — a parent or a tutor. The child has no account and no sign-in of their own. The chat platforms set their own age rules: Claude is for adults, and ChatGPT requires a parent's consent for teenagers. MathTrail is not aimed at children using a chat by themselves.

## What is stored, and where

### In your Google Drive

The profile is a single JSON file, `mathtrail-profile.json`, in a folder named `MathTrail` in your own Drive. You can open it, read it, copy it, move it or delete it like any other file, and Drive's bin and version history are yours.

It holds:

- a **pseudonym** for the child, the grade, their interests and anything you asked to avoid;
- an identifier for the child, so that the profile can be carried to a future edition;
- the interface language, if you chose one;
- the topic the lessons are kept to, if you or your child chose one;
- the country you live in and, in the United States, your state, if you chose to give them — they are there only to count how many families each country has, and you may leave them out;
- which topics are mastered, the rating numbers behind the difficulty, and a summary of past answers: per topic, when it was last given, how many attempts and successes there were, and which traps came up;
- where the ratings stood before each of the latest answers, and at the start of each of the last seven days with answers, so that the progress can show what moved since the last task and over the week;
- a counter of tasks accepted today, which is how the daily limit is kept;
- the current task, with its answer encrypted so that it cannot be read out of the file before the child answers.

It does not hold the child's name, birth date, school, address, photograph or anything else that identifies them. **The pseudonym is the only name involved, and it never appears inside a task.**

### On the server

Nothing. The service is a single program that answers a request and forgets it. It has no database, no disk and no cache of your data.

Your Google sign-in is not stored either: it is encrypted and placed inside the token the service hands to your chat, together with the country you signed in from, as described below. The key lives in Google Secret Manager and never leaves the server.

### In the logs

The service writes counts, not content: whether a task was accepted or rejected, why it was rejected, how long it took, how many attempts it needed, whether a limit was hit, and which version of the instructions was used.

To count how many children use MathTrail each day, week and month without knowing who any of them is, the lines about a task given, an answer and a topic mastered also carry:

- a **counting name** for the child: a code made, with a key of the service's own, from the profile's identifier and the calendar month. It stays the same for one month and is different the next, and nothing about the child — the identifier, the pseudonym, anything else — can be worked out from it;
- the chat (Claude, ChatGPT or another), the language of the task, the grade, the month the profile was made, and how many topics are mastered;
- the **country**, and the **state** in the United States, if you gave them in the profile, and the **country you signed in from**, each as a short code;
- for an answer, whether it was right, whether the hint was used, the **chance** of a right answer the service expected of its task, who chose the task (the service's rule, the chat or you), whether it was one of the first answers that find the level the child starts at, and, after those, the range the child's count of answers was in, such as 6 to 20.

Like every line about your sessions, these lines also carry an opaque code standing for your account: made with the service's own key from your Google account, it is neither your Google ID nor your email, and it changes at the first sign-in after that key is replaced. While the lines of two months are both kept, that code is what could join a child's two counting names.

These lines are kept for at most 62 days. What is kept after that are totals — how many children, how many tasks, in which countries: totals for each day, totals for each week and each month told apart by one thing at a time, such as the country or the grade, and, for each month, the traps that wrong answers fell into, by topic and grade, and how the answers to the tasks the service chose came out against the chance it expected, by that chance and by how many answers the child had given. They hold no identifier of any kind. A total counts a group of children, and a group may be a single child; a total of fewer than ten is never shown publicly.

The logs never contain the pseudonym, the text of a task, an answer, an email address, a token, the child's identifier, or the address a request came from.

Separately, Google Cloud Run keeps the standard request logs any service on it produces — time, path, response status and the address the request came from — for the platform's own retention period. Those hold no profile data, no task text and no answers.

### The country you sign in from

When you sign in, Google sends your browser back to MathTrail, and at that one moment the service looks up which country your browser's address belongs to, in DB-IP's free IP-to-Country database — a file the service keeps on its own server, so nothing is sent to DB-IP or anyone else. Only the two-letter code of the country is kept, sealed inside the token your chat holds; the address itself is neither stored nor written down. An address that belongs to no country, such as one of a private network, gives no country. Your chat's own requests come from its servers, not from you, so the country stays the one of your last sign-in.

You can have this country left out of what is counted: tick "Don't count the country I sign in from" in the form of the profile, or ask in the chat. From then on no line the service writes carries it, until you untick the box.

IP geolocation by [DB-IP](https://db-ip.com), under the [Creative Commons Attribution 4.0](https://creativecommons.org/licenses/by/4.0/) licence.

## Google sign-in and Drive

Signing in asks Google for two things:

- **`openid`** — to know that the sign-in succeeded;
- **`drive.file`** — the narrow Drive permission. It grants access **only to files this app itself created**. MathTrail cannot see, list or open anything else in your Drive, including files you made yourself.

Two consequences are worth knowing in advance, because they look like faults and are not:

- **Google asks for your permission every time you connect.** The service keeps no long-lived Google credential of its own, so each connection has to obtain one afresh, and Google shows its permission screen each time.
- **A Google account allows about a hundred of these grants at once.** After roughly a hundred connections the oldest one stops working, without a warning. Signing in again fixes it.

The token the service issues to your chat is valid for 15 minutes and is renewed automatically for up to 30 days at a time, and at most 90 days from the sign-in. After that you sign in again.

## What reaches the chat's model

When your chat asks for a task, MathTrail sends back the material the model needs to write one: the topic, the difficulty, example tasks, the formats to follow — and **the free-form notes you wrote about the child**, if you wrote any, because they are what lets the model pitch the wording at the right level.

Write those notes accordingly: they leave for the chat provider along with everything else in that message. The pseudonym does not go with them, and the rest of the profile — the ratings, the history, the answers — stays out of it.

The text of the task itself is written by the model and travels back through MathTrail's checks.

## What you can see, and what the chat shows

Chat hosts show the raw arguments of the calls an app makes. When the model submits a finished task, that submission — **including the correct answer** — is visible to anyone reading the tool-call log in the chat window.

MathTrail does not hide this and cannot. The answer is encrypted so that it cannot be taken from the profile file or handed out twice; it is not hidden from an adult who reads their own chat's log. If you want the task to be a surprise, do not scroll through the call log in front of the child.

## Advertising, profiling and training

There is no advertising in MathTrail. Nothing about your child is used to build an advertising profile, and nothing is sold or given to a data broker. MathTrail does not use your child's data to train any model.

What the chat provider does with the conversation is covered by their own policy — the same one that already applies to every other message you send in that chat.

## Cookies

This website sets no cookies, runs no analytics and loads nothing from other domains.

The service sets two cookies during sign-in only, both strictly necessary: one that ties the start of a sign-in to its finish, and one that remembers which apps you have already approved so you are not asked twice. Neither is used for tracking.

## Taking your data back, and deleting it

Everything is in your own Drive, so all of it is in your hands:

- **Export** — open the `MathTrail` folder in Drive and download `mathtrail-profile.json`. That file is the whole profile.
- **Delete** — delete the file. Empty Drive's bin to remove it completely. Nothing of the profile survives elsewhere.
- **Cut off access** — go to your Google account's [linked apps](https://myaccount.google.com/linkedapps) and remove MathTrail's access to your account. The service can then reach nothing.
- **Remove the app** — disconnect MathTrail in your chat's own settings.

Deleting the file and revoking access are independent, and doing both leaves nothing behind.

## Who else is involved

- **Google** — the sign-in and Drive, where your file lives.
- **Google Cloud** — where the service runs (Cloud Run), and where the totals above are kept (BigQuery).
- **GitHub Pages** — where this website is served from; GitHub records its own access logs for it.
- **Your chat provider** — Anthropic for Claude, OpenAI for ChatGPT.

There is nobody else.

## Changes

If this policy changes, the date at the top changes with it and the previous wording stays in the project's public history.

## Languages

This policy is published in several languages. The English text is the authoritative one; where a translation disagrees with it, the English version applies.
