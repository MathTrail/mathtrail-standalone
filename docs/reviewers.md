# Reviewing MathTrail

MathTrail is a free, open-source app for olympiad-style maths, grades 1 to 6, that runs inside Claude and ChatGPT. An adult — a parent or a tutor — runs the lesson in their own chat, and the child answers on a card beside them. The chat's model writes each task; MathTrail chooses what comes next, checks the task with a program before the child sees it, records the answer and keeps the child's ratings.

This page is for a directory's reviewer: how to connect, how to sign in with the account given in the review form, what to try and what to expect. The password is in the review form, and nowhere else.

## Connect

The address is `https://mcp.mathtrail.app/mcp`.

- **Claude:** open [Customize → Connectors](https://claude.ai/customize/connectors), press **+**, choose **Add custom connector**, paste the address, press **Add**, then **Connect**.
- **ChatGPT:** connect the app as the review describes, or, in developer mode, add an app with the address above.

## Sign in

Connecting opens MathTrail's own page, "Connect … to MathTrail?".

1. At its foot, open **Reviewing MathTrail for a directory?**
2. Type the reviewer's password from the review form, and press **Sign in as the reviewer**.
3. The browser goes back to the chat, connected to the demo profile. Google is not involved: there is no Google sign-in, no code and no confirmation step.

Do not press **Allow** on that page. It signs a parent in with Google, and the browser remembers it: the next time, it goes straight to Google and the reviewer's line is not shown. If that happens, clear the cookies of `mcp.mathtrail.app` and connect again.

## The demo profile

The account opens one demo profile: a child known by a pseudonym, with the trial series done and about twenty tasks answered, some of them wrongly on purpose. It is a profile made for review, kept in a demo Google account's Drive, and nothing done in it reaches a real family or is counted as a child.

## What to try

In a chat with MathTrail switched on, type as the adult would; the card is the child's part:

1. **A task.** Say "Let's do a MathTrail task." A card shows that the task is being written; the model takes a minute or two. Then the card shows the task and five options, and the model goes on to write the next task ahead, saying nothing about it.
2. **An answer.** Choose an option on the card, as the child would, or tell the chat the child's answer. An option chosen on the card turns the card into how the answer went at once, with nothing sent to the chat: the mistake behind a wrong answer, a picture of the solution, and the solution step by step. An answer told in the chat gets the same card below. "We don't know this one", said in the chat, is an answer too.
3. **The hint.** Press the bulb, **Hint**, on the card before answering.
4. **Another task.** Press the two arrows in a circle, **Another task**, on a task's card before its answer, or **Another task** under how the answer went. The card puts its words in the chat — in Claude, press Enter to send them —, the task written ahead appears at once on a new card below, and the model writes the one after it. Or ask for one in the chat, or ask for a topic: "Let's practise counting."
5. **The progress.** Say "Show the progress." The card shows each topic's rank, what moved since the last task and over the past week, the **Review** — strengths, topics to develop and what to do next — and **Mistakes that repeat**.
6. **The profile.** Open **Profile and progress** at the top of a task's card: its **Profile · for the parent** section shows the profile, and **Edit** there changes the interests or the grade. Or ask in the chat: "Change my child's interests to space and dinosaurs." The profile has no card of its own: asked to show it, the chat tells it in words.

## Where the answer is shown

The card never carries the right answer before the child has answered: MathTrail keeps it sealed, and opens it only once the answer is recorded.

The chat's own log of tool calls is another matter. When the model hands its task to MathTrail, that call carries the whole task with its right answer, and Claude and ChatGPT show a call's arguments as they are. A reviewer who opens the log sees the answer; a parent is told, in the privacy policy, not to read that log with the child beside them.

## Limits

- **50 new tasks a day** for the demo child, shared by every reviewer, counted until midnight by the clock of the device a card was last answered on, or midnight UTC before any card has been answered. Past them the card wishes the child a good tomorrow and says the day's tasks are over; the progress and the last task stay available.
- **A task is written by the chat's model** and checked before it is shown, so it takes a minute or two; one the checks refuse is written again, up to three times.

## Contact

[altedtech.info@gmail.com](mailto:altedtech.info@gmail.com). The code is at [github.com/MathTrail/mathtrail-standalone](https://github.com/MathTrail/mathtrail-standalone), and the privacy policy at [mathtrail.app/en/privacy/](https://mathtrail.app/en/privacy/).
