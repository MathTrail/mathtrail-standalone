---
title: Help
description: How to connect MathTrail, how a lesson goes, what to do when something looks wrong, and how to reach us.
---

# Help

What you need, how to connect MathTrail, how a lesson goes, what to do when something looks wrong, and how to reach us.

## What you need

- A Claude or ChatGPT account of your own, and a Google account. Claude is for adults, and ChatGPT needs a parent's consent for teenagers.
- Your child needs neither: they have no account and sign in nowhere.

MathTrail keeps your child's profile in one file in your own Google Drive, and nowhere else.

## Connect it in Claude

1. In Claude, open [Customize → Connectors](https://claude.ai/customize/connectors).
2. Press **+** and choose **Add custom connector**.
3. Paste the address: `https://mcp.mathtrail.app/mcp`
4. Press **Add**, then **Connect**, and sign in with Google. MathTrail asks to keep one file with your child's profile in your Google Drive. Allow it on Google's screen: that file is where your child's progress is saved.
5. In a chat, press **+** → **Connectors**, switch MathTrail on and say: "Let's do a MathTrail task."

Claude's free plan allows one custom connector. A connector added on claude.ai or in Claude Desktop works in Claude's mobile apps too, on the same account.

## Connect it in ChatGPT

MathTrail is not in ChatGPT's directory yet. Until it is, ChatGPT's developer mode adds it by its address on paid plans, on the web: turn developer mode on and add an app with the address above, as [OpenAI describes](https://developers.openai.com/plugins/deploy/connect-chatgpt). MathTrail is still being tested in ChatGPT.

## How a lesson goes

- **The first time**, the chat asks you to say you are the child's parent or tutor, then for a pseudonym for your child, never a real name, and the grade from 1 to 6. If you like, add their interests and anything to leave out of the tasks.
- **The first five tasks** find where your child stands. After them, the tasks follow the answers.
- **You type, and your child taps.** Ask for a task in the chat. A card appears, and in a minute or two the task arrives on it with five options; your child answers by tapping one. **The bulb** opens a hint. **The two arrows** ask the chat for the next task, and once your child has answered, **Another task** does the same. The task comes below, in a new card. While your child solves, the chat writes the next task ahead, so it usually appears on the new card at once.
- **After an answer** the card at once turns into how the answer went: the mistake behind a wrong answer, a picture of the solution, and the solution step by step. Nothing goes to the chat. If your child does not know, tell the chat so, and a card of the solution comes below.
- **Progress:** ask "How is my child doing?" to see where your child stands in each topic, what moved over the past week, a review and the mistakes that repeat. **Edit** there changes the profile, or ask in the chat.
- **Keep the chat's log of tool calls closed while your child solves:** it shows each task together with its answer.

## When something goes wrong

- **Google sends you back to a MathTrail page, "Tick the box for Google Drive".** The box that lets MathTrail keep its file was left unticked. Press **Back to Google** and tick it: without that file there is nowhere to keep the progress.
- **MathTrail asks you to sign in again.** A sign-in lasts at most 90 days. A Google account also allows about a hundred grants to one app at a time, so after many connections the oldest one stops working. Signing in again fixes both.
- **Claude will not add another connector.** Its free plan has room for one custom connector: remove another one first.
- **The card waits a minute or two.** The chat's model writes each task, and MathTrail checks it before your child sees it. The first task of a lesson takes that long; the next ones are written while your child solves, and come at once when they are ready. The chat's usual depth of thinking keeps the wait short: at its deepest, the model thinks over each task for a minute or more. A task the checks refuse is written again, with three attempts in all; if all three fail, the chat says so, and you can ask for another.
- **The card says "There are no more new tasks today".** A child gets up to 20 new tasks a day, and the day also closes after five requests in which no task could be written. The day starts at 00:00 UTC. The progress and the last task stay available.
- **Claude puts "Another task" in the message box under "Use caution before running this prompt".** That is the card's request waiting for you: press Enter to send it.
- **Claude asks whether to allow MathTrail to save the profile.** Saving replaces what the profile held before, so Claude asks first. Allow it when you asked for the change.
- **No card appears.** Some chats show no cards. The chat then gives the task in words with its options A to E, and you type the letter of your child's answer.
- **The profile cannot be read, or it is in Drive's bin.** MathTrail says so in the chat and changes nothing. Restore the file from the bin, or let MathTrail put back its latest readable version or start a new profile, as [the guide to your child's data](https://github.com/MathTrail/mathtrail-standalone/blob/main/docs/parents.md) explains.
- **A task looks wrong.** It may be: the program checks what it can, and the meaning is checked by the model itself. Tell us what the task said.

## Your child's data

Everything MathTrail knows about your child is one file, `mathtrail-profile.json`, in a `MathTrail` folder of your Google Drive. How to download it, disconnect MathTrail or delete everything is in [the guide to your child's data](https://github.com/MathTrail/mathtrail-standalone/blob/main/docs/parents.md). What is kept, where and for how long is in the [privacy policy](../privacy/).

## Contact us

- **Help, and questions about your child's data:** write to [altedtech.info@gmail.com](mailto:altedtech.info@gmail.com).
- **A mistake in a task, or an idea:** write to the same address, or open an issue on [GitHub](https://github.com/MathTrail/mathtrail-standalone/issues). Issues are public, so put no name, photograph or other detail of your child in one.
- **A security problem:** report it privately through [GitHub's security advisories](https://github.com/MathTrail/mathtrail-standalone/security/advisories/new), as the project's [security policy](https://github.com/MathTrail/mathtrail-standalone/blob/main/SECURITY.md) asks.

We read everything and answer as soon as we can, but there is no promised response time.
