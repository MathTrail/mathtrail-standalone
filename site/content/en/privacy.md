---
title: Privacy policy
description: What MathTrail knows about your child, where it is kept, and how to take it back or delete it.
---

# Privacy policy

Last updated: 21 September 2026.

MathTrail is a free, open-source project. Questions about this policy or about your child's data go to [alexander.ryazanov.1987@gmail.com](mailto:alexander.ryazanov.1987@gmail.com).

## The short version

- **We store nothing.** The service keeps no database and no files. Between one request and the next it remembers nothing.
- **Your child's profile lives in your own Google Drive**, in a file you can open, copy or delete at any moment.
- **A pseudonym, never a real name.** No birth date, no school, no photograph.
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
- which topics are mastered, the rating numbers behind the difficulty, and a summary of past answers: per topic, when it was last given, how many attempts and successes there were, and which traps came up;
- a counter of tasks accepted today, which is how the daily limit is kept;
- the current task, with its answer encrypted so that it cannot be read out of the file before the child answers.

It does not hold the child's name, birth date, school, address, photograph or anything else that identifies them. **The pseudonym is the only name involved, and it never appears inside a task.**

### On the server

Nothing. The service is a single program that answers a request and forgets it. It has no database, no disk and no cache of your data.

Your Google sign-in is not stored either: it is encrypted and placed inside the token the service hands to your chat. The key lives in Google Secret Manager and never leaves the server.

### In the logs

The service writes counts, not content: whether a task was accepted or rejected, why it was rejected, how long it took, how many attempts it needed, whether a limit was hit, and which version of the instructions was used.

The logs never contain the pseudonym, the text of a task, an answer, an email address or a token.

Separately, Google Cloud Run keeps the standard request logs any service on it produces — time, path, response status and the address the request came from — for the platform's own retention period. Those hold no profile data, no task text and no answers.

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
- **Cut off access** — go to your Google account's [third-party access settings](https://myaccount.google.com/connections) and remove MathTrail. The service can then reach nothing.
- **Remove the app** — disconnect MathTrail in your chat's own settings.

Deleting the file and revoking access are independent, and doing both leaves nothing behind.

## Who else is involved

- **Google** — the sign-in and Drive, where your file lives.
- **Google Cloud Run** — where the service runs.
- **GitHub Pages** — where this website is served from; GitHub records its own access logs for it.
- **Your chat provider** — Anthropic for Claude, OpenAI for ChatGPT.

There is nobody else.

## Changes

If this policy changes, the date at the top changes with it and the previous wording stays in the project's public history.

## Languages

This policy is published in several languages. The English text is the authoritative one; where a translation disagrees with it, the English version applies.
