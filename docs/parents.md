# Your child's data: taking it, disconnecting, deleting

For the parent or tutor who connected MathTrail. Everything MathTrail knows about your child is in one file in your own Google Drive, so everything on this page you do yourself — in your Drive, your Google account and your chat — without asking anybody and without waiting.

## Where it is

In your Google Drive, in a folder named **MathTrail**, the file **`mathtrail-profile.json`**. That one file is the whole profile: the pseudonym, the grade, the interests, the ratings, a summary of past answers and the task on the card. Ask the chat where your child's profile is, and MathTrail gives it the folder, the file and a link; the *Profile and progress* card shows where it is, under *Your data*.

The service keeps nothing between requests: no database, no copy of the file, and no pseudonym, task or answer in its logs. What the file holds, field by field, is in the [privacy policy](https://mathtrail.app/en/privacy/).

Two more kinds of file can turn up in the folder:

- **`mathtrail-profile set aside <date>.json`** — a profile that could not be read, put aside when you asked for a new start in its place. MathTrail never reads it again; it is left there so that nothing is lost by mistake. Delete it when you no longer need it.
- **Another file with a profile in it.** MathTrail reads only the newest, and the *Profile and progress* card names the others, so that you can delete them.

## Take it with you

Download `mathtrail-profile.json` from Drive like any other file. That is the whole export: plain JSON, readable in any text editor. One part of it is sealed — the answer, the solution and the explanations of the task on the card, encrypted so that nobody reads them before the child answers.

Earlier states of the file are in its version history in Drive, and any of them can be downloaded from there. MathTrail asks Drive to keep the first version of each day the profile changed, and lets at most a hundred kept versions stand, the earliest going first.

## Disconnect MathTrail

There are two switches, independent of each other. Neither one touches the file.

**In your chat**, so that the chat stops using MathTrail:

- **Claude** — open [Customize → Connectors](https://claude.ai/customize/connectors), find MathTrail and remove it from its **⋯** menu. To switch it off in one conversation only, use **+** → **Connectors** in that conversation.
- **ChatGPT** — in ChatGPT's settings, find MathTrail among the apps and disconnect it, as [OpenAI describes](https://help.openai.com/en/articles/20001494).

**In your Google account**, so that MathTrail can no longer reach your Drive: open your account's [linked apps](https://myaccount.google.com/linkedapps), choose **Access to your Google Account**, then MathTrail, **See details** and **Remove access**, and confirm. Google's own steps are in [its help](https://support.google.com/accounts/answer/13533235). This stops MathTrail in every chat at once, Claude and ChatGPT alike: the next request asks you to connect it again.

Connecting again later, with the same Google account, finds the same profile: it lives in your Drive, not in the chat.

## Delete everything

1. In Drive, delete `mathtrail-profile.json` — or the whole `MathTrail` folder, with any files set aside in it. They go to Drive's bin (Trash).
2. Empty the bin, or delete the files from it forever. Drive deletes them from the bin by itself after 30 days.
3. Disconnect MathTrail as above, unless you mean to start again: while it is connected, the next lesson begins by asking you for a new profile.

A file deleted forever is gone for good, together with its version history: nobody can bring it back, MathTrail included, because there is no other copy anywhere. While the file is still in the bin, MathTrail says so when it is asked for the profile: restore the file in Drive and everything is as it was, or ask the chat to start a new profile.

## What else keeps something

- **Your chat.** The conversations of the lessons — the tasks, your child's answers and the explanations — are in your chat history, kept by Anthropic or OpenAI under their own policies. Disconnecting MathTrail does not delete them: delete those conversations in the chat if you want them gone. The chat's log of tool calls also shows each task as the model handed it in, with its answer, so read it after the lesson rather than in front of your child.
- **Your Google account** lists MathTrail among its linked apps until you remove its access.
- **MathTrail's logs** hold counts and timings, and an opaque code standing for your account: made from it with the service's own key, it is neither your Google ID nor your email, and it changes at the first sign-in after that key is replaced. Never the pseudonym, a task or an answer. What the platform records about each request is set out in the [privacy policy](https://mathtrail.app/en/privacy/).

## Another Google account

A different Google account has a Drive of its own, and MathTrail starts a new profile there. The first one stays where it was; sign in with the first account again to find it.

Questions about your child's data go to the address in the [privacy policy](https://mathtrail.app/en/privacy/).
