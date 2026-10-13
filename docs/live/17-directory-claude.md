# Claude's directory: the submission (T87)

**Submitted on 2026-10-13, and in review: a reviewer is checking MathTrail's listing.** The submission went from the author's own Claude account, on the Max plan, through the developer portal's ten steps, against the service at v0.6.4 (`014e8a4`). The portal reads "In review: A reviewer is checking your listing. We'll email your primary contact if we need anything."

## What ran

- **The portal.** [claude.ai/directory/manage](https://claude.ai/directory/manage), **Submit new**, then **MCP connector**: ten steps, as [Submit a connector to the directory](https://claude.com/docs/connectors/building/submission) describes them (read 2026-10-13), in the evening of 2026-10-12 by the author's clock.
- **The service.** v0.6.4 at `014e8a4`. The first attempt ran against v0.6.2, whose Tools step flagged `submit_task` (finding 1). T107 typed the parameters and went out in v0.6.3, and T108 mended the lift the CI of that merge caught (finding 2), in v0.6.4. The second attempt ran against v0.6.4 and was submitted.
- **Before it.** The demo profile refreshed and the reviewer's tour run the same evening ([report 14](14-reviewer-access.md), T82). The listing's pictures were taken again from the card of v0.6.2 with `just listing-shots` and looked at; the portal asked for none (finding 4).

## The steps, as entered

1. **Connection.** `https://mcp.mathtrail.app/mcp`, one address for every user. The portal read 9 tools, 1 resource and OAuth from the live server.
2. **Tools.** The 9 tools the model sees, grouped as read-only (`get_package`, `get_profile`, `get_progress`, `show_result`) and writing (`next_task`, `prepare_task`, `submit_answer`, `submit_task`, and `save_profile`, marked destructive), none flagged in the second attempt. The review page counts 12, the three that only the card calls among them.
3. **Listing.** The name, the one-liner, the description and the category Education, from [listing.md](../listing.md); the documentation, the privacy policy, the support address and the slug `mathtrail`. The icon is asked as an address and checked with **Check icon**. The portal recommends the favicon of the server's address, which Claude takes from Google and which still showed the earlier mark, so the icon given is the logo of 1024 pixels at `https://raw.githubusercontent.com/MathTrail/mathtrail-standalone/main/plugin/assets/logo.png`, which the preview then showed.
4. **Use cases.** Up to three pairs of a use case and an example prompt, the prerequisites, and whether the connector reads, writes or both: the listing's three examples as pairs, its prerequisites, and Read and write. With Read only chosen, the portal named the five tools whose annotations write.
5. **Company.** ALT EdTech and `https://mathtrail.app`. The primary contact is the author, at the address of the Claude account, which the portal fills in; the listing's public support contact is `altedtech.info@gmail.com`.
6. **Authentication.** Detected: "sign-in required (OAuth with a Client ID Metadata Document)", with the server answering 401 to the connection and 200 for the sign-in's settings and its provider. Entered as detected: always required, with a client ID metadata document.
7. **Data handling.** We own the API; no personal health data; no sponsored content.
8. **Test & launch.** One field for the reviewers' instructions, and none of its own for credentials: the instructions of [listing.md](../listing.md), with the reviewers' password on their last line, from the password manager. "I have run every tool via MCP Inspector or as a custom connector in Claude" ticked: every tool had been run through the author's custom connector in Claude, in T82 and on 2026-10-13, `edit_profile` through the profile's form.
9. **Compliance.** The seven acknowledgments, all ticked (finding 5), and the additional note of [listing.md](../listing.md).
10. **Review and submit.** No quality warnings, and "Review can take a few weeks, and approval isn't guaranteed." Then "Server submitted for review".

## Findings

1. **The Tools step asks every parameter for a type.** Against v0.6.2 it flagged `submit_task`: "To fix: Add a type to these parameters: brief, task, self_check." The three take any JSON (R100, R137), and no page of the documentation asked for a type. T107 typed a value of any JSON with every kind of JSON (R304), and the second attempt flagged nothing.
2. **The CI of T107's merge failed on two jobs.** The runtime image: gcr.io answered "not found" for the pinned distroless digest, which resolved again within the hour as the current `latest`, and a re-run passed. The tests: the gopter property of the lift found a topic left 1.8e-16 below its target, a defect of T103 that T108 mended before the submission.
3. **A draft lives in its browser tab.** The portal says "Draft saved in this tab": once the tab was closed, the first attempt's steps were gone, and the list of submissions showed nothing.
4. **No screenshots were asked for.** The pictures `just listing-shots` takes stay ready in `web/listing/`, for a later edit of the listing.
5. **The seven acknowledgments**, as the portal words them:
   1. "I have read and understand the MCP Directory developer guidelines."
   2. "My server calls our own first-party APIs (or APIs we legitimately proxy with permission)."
   3. "My MCP server does not transfer money, cryptocurrency, or other financial assets on a user's behalf."
   4. "My MCP server does not use AI models to generate images, video, or audio."
   5. "Tool descriptions contain no instructions about model behavior, other tools, or external instruction sources, and no hidden or encoded text."
   6. "My MCP server does not collect conversation data beyond what is needed for its function."
   7. "Public documentation for this MCP server will be live by the publish date."
6. **The listing's page in the portal** shows each tool with its label, Read-only, Writes or Destructive, "declared by this server, in its own tool list dated October 12, 2026", and "Works with: Claude (web & mobile), Claude API, Claude Code, Claude Desktop". Health, usage and each tool's calls start once the listing is live.

## Next

- The status, and any word from a reviewer, which goes to the primary contact. A listing a person approves goes live once the author presses **Publish**.
- T106: the demo profile refreshed by 2026-10-19, and again within six days of each refresh, while the review lasts.
- After publication: MathTrail connected from the directory on the free plan, on the web and on a phone, the whole lesson, and whether Claude's warning on the card's message stays (CL-42); the site, the README and the page Help linking to the listing; the statuses in [catalogs.md](../catalogs.md); the rest of this report.
- The reviewers' password changed once the review is over.

## Cost

- $0 in the cloud, the author's Claude plan, and no chat messages for the submission itself.
