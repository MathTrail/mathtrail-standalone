# The Claude and ChatGPT directories: what they ask, and where MathTrail stands

As of 5 October 2026. Every page cited here was read on that day; a row that rests on a search summary, because the page refused to be fetched, says so. This is what phase 12 of the plan works from: everything each directory asks of a submission and of the service, and, against each requirement, where MathTrail already meets it, which task closes the gap, or which question is the author's. Tasks T71.2–T71.4 take their lists from section 12. The section numbers stay as they are, since T71.4 adds the listing texts to this document. T71.2 closed its rows the same day: each says so in its status, with the decision that did it, and its finding stays what the snapshot found.

**How to read it.** Section 1 holds the questions the phase waits on, each with a recommendation, and sections 2–6 give each of them its facts. Sections 7 and 8 are the requirements row by row and the gaps they leave. Section 10 is what only a submission, a publication or a live run can tell.

**The status of a requirement:**

- **Met** — MathTrail meets it today, and the row says where;
- **Gap → T71.x** — a task of phase 12 closes it;
- **Met in T71.x** — a task of phase 12 closed it, by the decision named;
- **Decision → О-n** — it waits on the author's answer to a question in PRODUCT-V1 12.2;
- **Unknown → §10** — only a submission, a publication or a live run tells;
- **N/A** — it does not apply, or it is a fact to plan by rather than something to do.

The service is at `https://mcp.mathtrail.app/mcp`, and its code is cited by file and line. PRODUCT-V1 is the author's working document and is not kept in the repository; section 1 restates every question it holds for this phase.

## 1. Decisions for the author

| Question | Recommendation | What waits on it |
|---|---|---|
| **О-61. The child in an adult's chat** (§2). Both platforms' terms forbid making an account available to anyone else; Claude is for adults and disables accounts on "indicators of minor activity"; a ChatGPT plugin may not target children under 13. MathTrail's instructions tell the model to talk with the child. | **B and C together:** rewrite the instructions and the tools' descriptions so that the model talks to the adult while the child answers on the card beside them, and send the two letters of appendix A as soon as this and О-63 are answered, describing that lesson and that card, so that the replies come while the rewrite is made. | T71.4's texts, T71.5, T71.6 |
| **О-62. How a reviewer signs in** (§3). | **B:** a sign-in of MathTrail's own for one named reviewer account, off unless its secret is set, reaching only the demo profile. | T71.3 |
| **О-63. The card against the hosts' design rules** (§4). | A widget task before T71.4: the progress opens full screen from the task card rather than replacing the task inside it; the choice of a topic moves into that progress, so that the row under a task keeps two actions, "Hint" and "Another task"; and the mark leaves the card's header in ChatGPT, which draws the app's logo and name itself. The five answer options stay, explained as one choice. | T71.4's screenshots, T71.5 |
| **О-64. The Coach page, two steps from the card** (§5). | Take "Coach" out of the menu and link the page from "About"; the page keeps its prototype and its price. | T71.4, T71.6 |
| **О-65. Who submits on each platform** (§6). | One identity on both: the company, if it can pass OpenAI's business verification now; otherwise the author in person, knowing that ChatGPT shows that name. | T71.4, T71.5, T71.6 |
| **О-66. A version in the widget's address** (SPEC remark 38). ChatGPT keeps a page by its address for up to an hour and asks that every published address keep working. | Keep the one fixed address, and keep each change of the card readable by the page of the release before for an hour; a payload a page cannot read already shows as one it cannot show (R135). | T71.2 |

Not questions, because a task of the phase owns them: the tools' annotations (T71.2 decides them by an R entry, after the live check of §10), the timestamps and identifiers in results (T71.2), the widget's sandbox domain (T71.2), and the support and documentation pages (T71.4). The ChatGPT sign-in needs no `UserInfo` (GPT-42), so the audit's question about an email address is dropped.

## 2. The child in an adult's chat

### 2.1 What the platforms say

| Platform | Rule | Source |
|---|---|---|
| Claude | "You must be at least 18 years old or the minimum age required to consent to use the Services in your location, whichever is higher." | [Consumer Terms](https://www.anthropic.com/legal/consumer-terms), effective 8 October 2025, read 2026-10-05 |
| Claude | "You may not share your Account login information … You also may not make your Account available to anyone else." | The same |
| Claude | "We have safety systems in place to detect if people under 18 may be using Claude and we'll disable accounts based on indicators of minor activity." An account comes back after an age check with Yoti. | [Age assurance on Claude](https://support.claude.com/en/articles/15171100-age-assurance-on-claude), 18 May 2026, read 2026-10-05 |
| Claude | In some US states the mobile app takes the app store's age signal and blocks anyone under 18. | [Minimum age requirement](https://support.claude.com/en/articles/13117299-minimum-age-requirement-access-restriction), 16 March 2026, read 2026-10-05 |
| Claude | "Products serving minors, including organizations providing minors with the ability to directly interact with products that incorporate our API(s), must comply with the additional guidelines outlined in our Help Center article" — age verification, moderation, a child-safety system prompt, COPPA and disclosure. | [Usage Policy](https://www.anthropic.com/legal/aup), effective 15 September 2025; [Guidelines for organizations serving minors](https://support.claude.com/en/articles/9307344-responsible-use-of-anthropic-s-models-guidelines-for-organizations-serving-minors), 16 March 2026; both read 2026-10-05 |
| Claude | "Our first-party consumer services are restricted to users 18 years and older"; "if your product allows minors to interact directly with our models, please refer to our Guidelines for Organizations Serving Minors". | [Child safety guidance for developers](https://support.claude.com/en/articles/15591275-child-safety-guidance-for-developers), 26 June 2026, read 2026-10-05 |
| Claude | "Software must not violate or facilitate violation of our Usage Policy." The directory's rules name no age. | [Directory Policy](https://support.claude.com/en/articles/13145358-anthropic-software-directory-policy) 1.A, 15 April 2026, read 2026-10-05 |
| ChatGPT | "You must be at least 13 years old or the minimum age required in your country to consent to use the Services. If you are under 18 you must have your parent or legal guardian's permission to use the Services." | [Terms of Use](https://openai.com/policies/row-terms-of-use/), search summary: the page refused fetching, 2026-10-05 |
| ChatGPT | "You may not share your account credentials or make your account available to anyone else and are responsible for all activities that occur under your account." | The same, search summary |
| ChatGPT | "Plugins must be suitable for general audiences, including users aged 13–17. Plugins may not explicitly target children under 13." | [Plugin guidelines](https://developers.openai.com/plugins/plugin-guidelines), read 2026-10-05 |

Neither directory's documents mention a parent using an app with a child.

### 2.2 What MathTrail says

| Where | It says |
|---|---|
| PRODUCT 3 (`PRODUCT-V1.md:66–67`) | The adult "owns the chat account" and is "the user from the platforms' point of view"; the child "solves tasks in the parent's chat, next to them or with their permission". |
| The instructions (`content/instructions/mcp_instructions.md:3`) | "You coach one child through short olympiad-style maths tasks … you talk with the child and write the tasks." |
| The instructions, lines 14, 36, 41 and 49 | "only when the child asks for another"; "tell the child this one did not work out"; "a child who does not know says so in the chat"; "Another task" sends "those words to the chat as the child's message". |
| The instructions, line 23 | Before a profile: "first ask the adult to say they are the child's parent or tutor". |
| Counts | "child" 39 times on 23 of the instructions' 57 lines, "adult" 9; in the ten tools' descriptions, "child" 27 times and "adult" 6. |
| A tool's description (`internal/transport/mcp/submittask.go:158–165`) | "never say which option is right before the child has answered, whatever the child asks." |
| README (`README.md:60`) | "MathTrail is used by an adult — a parent or a tutor — in their own chat …, and the child solves the tasks next to them." |
| The privacy policy (`site/content/en/privacy.md:29`) | "MathTrail is not aimed at children using a chat by themselves." |
| The terms (`site/content/en/terms.md:22`) | "The adult runs the lesson." |
| The grants package (`docs/grants/README.md`) | "An adult, a parent or a tutor, runs the lesson in their own chat, and the child answers on a card in it." |

### 2.3 Where they clash

- The documents describe a lesson an adult runs. The text the model and a reviewer read describes a model talking with a child who types in the chat. Both platforms' terms forbid making an account available to anyone else, and a child typing "I don't know" or "another one" is what Claude's age assurance looks for.
- PRODUCT 3 lets a child use the chat "with their permission", alone. Neither platform allows that for a child of grades 1–6: Claude admits nobody under 18, ChatGPT nobody under 13.
- ChatGPT's rule is about whom the plugin targets. MathTrail is positioned for parents and tutors (О-17), but its tasks are for grades 1–6 and its descriptions speak of the child throughout.
- Anthropic's guidelines for products serving minors are written for products built on its API, which MathTrail is not; its child-safety page sends a product "that allows minors to interact directly with our models" to them all the same.

### 2.4 The options

| | A. Keep it, and describe a lesson the adult runs | B. The model talks to the adult; the child answers on the card | C. Ask the platforms first |
|---|---|---|---|
| What changes | The listing's and the reviewer's texts | The instructions and the tools' descriptions; the words "Another task" sends, sent as the adult's; what the child says in the chat today — "I don't know" (R208) and a question about the task (R145) — said by the adult; PRODUCT 3's "or with their permission"; the tests that hold the instructions' sections | Nothing until the replies; the letters are appendix A |
| Cost | Nothing beyond T71.4 | A task of its own before T71.4, and a lesson of some ten tasks in Claude to check it — a few dozen messages of a subscription's limit | The wait: Anthropic's escalations go to `mcp-review@anthropic.com`, OpenAI answers through its support, and neither promises a time |
| Risk | The text a reviewer reads contradicts the description; a family's account may be disabled when the child types in the chat | The lesson is less direct for the child, and the model may still answer the child, which the rewrite and its tests have to prevent | No reply, or a general one; the phase waits meanwhile |

**Recommendation: B, with C's letters sent as soon as this question and О-63 are answered.** B makes the text the model reads say what the product's own documents already say, and it takes the child out of the chat's input, which is what both platforms' terms and Claude's age assurance turn on. The letters describe B's lesson and ask whether it is acceptable; the letter to Anthropic also asks about the package (CL-18). The submission starts once B is done, and a reply that rejects even B stops the phase.

## 3. How a reviewer signs in

Every call to `/mcp` needs a sign-in with Google and `drive.file`. There is no demo mode, and the development sign-in refuses to start on a deployment (SPEC 9.4, R82). What the platforms and Google say:

- ChatGPT's reviewer account "should work immediately without MFA approval, email or SMS codes, magic links, or private-network access" ([Submission](https://developers.openai.com/plugins/deploy/submission), read 2026-10-05), and "Plugins that require additional login steps, such as a new account sign-up or 2FA through an inaccessible account, will be rejected" ([Plugin guidelines](https://developers.openai.com/plugins/plugin-guidelines), read 2026-10-05).
- Claude asks for "credentials for a fully populated account" ([Checklist](https://claude.com/docs/connectors/building/review-criteria), read 2026-10-05).
- Google: "Even if you haven't turned either of these settings on, Google might also ask you to tap a notification to help confirm it's you signing in", and "If we notice something different about how you sign in, like your location, we may ask you to take extra steps to confirm it's you" ([Sign in with Google prompts](https://support.google.com/accounts/answer/7026266?hl=en), read 2026-10-05).

| | A. A Google account without 2-Step Verification | B. A reviewer sign-in in MathTrail's authorization server |
|---|---|---|
| How | The reviewer signs in to Google with the demo account's password, as a parent does | On the consent step, a password for one named reviewer account in place of Google; the server holds the demo account's Google refresh token and signs the reviewer in as that account |
| Code and review | None | A way in on the deployment: a plan, tests and a review in T71.3, which its text already allows |
| SPEC and 02-auth | Nothing | 9.1 a form or an endpoint; 9.3 its parameters; 9.4 and R82, "nobody on a deployment", become one named account behind a secret; 11.2 two variables; 02-auth a threat, guessing, held to the address's pace |
| Money | $0 | $0: two secrets beside today's three, within Secret Manager's six free active versions a month, $0.06 a version past them ([pricing](https://cloud.google.com/secret-manager/pricing), read 2026-10-05) |
| Google's check | Likely: reviewers sign in from their companies' networks and devices, often in another country, and a check is the "additional login step" ChatGPT rejects | None: Google is not asked |
| Re-reviews | Each one signs in to Google again, with the same chance of a check | It works while the secret and the Google refresh token live. Google ends a refresh token "not used for six months", and one of an app still in Testing after seven days ([OAuth 2.0](https://developers.google.com/identity/protocols/oauth2), read 2026-10-05), so the app must be In production, which T71.3 confirms |
| If the credentials leak | The demo Google account | The demo profile; the password is changed after each review |
| Either way | The demo profile lives in the demo account's Drive, and its 20 tasks a day (`MATHTRAIL_DAILY_TASKS`) are shared by every reviewer, which T71.3 counts | |

**Recommendation: B**, off unless its secret is set. ChatGPT names extra login steps as a reason to reject, Google says it may add one without 2-Step Verification, and a rejection costs a review cycle of unknown length.

## 4. The card against the hosts' design rules

Claude makes its guidelines binding: developers "must … follow design guidelines Anthropic publishes applicable to Software" ([Directory Policy](https://support.claude.com/en/articles/13145358-anthropic-software-directory-policy) 3.H, read 2026-10-05). ChatGPT's are written for "optional plugin UI", and its plugin guidelines ask to "review" them ([UI guidelines](https://developers.openai.com/plugins/concepts/ui-guidelines), [Plugin guidelines](https://developers.openai.com/plugins/plugin-guidelines), both read 2026-10-05).

| Element of the card | Claude ([design guidelines](https://claude.com/docs/connectors/building/mcp-apps/design-guidelines), read 2026-10-05) | ChatGPT ([UI guidelines](https://developers.openai.com/plugins/concepts/ui-guidelines), read 2026-10-05) | Recommendation |
|---|---|---|---|
| Five answer options, each press a call of `submit_answer` (`web/src/widget/TaskCard.tsx:290–298`) | "Direct manipulation like sliders, toggles, and selections" belongs in the app; an inline card has "At most 2 actions" | "Actions should perform either a conversation turn or a tool call", so each option is one | Keep, and say in the submission that they are one choice |
| "Hint" and "Another task" (`web/src/widget/LessonFoot.tsx:25–34`) | "At most 2 actions, placed at the bottom of the card" | "Limit to two actions, placed at bottom of card" | Keep: two actions at the bottom |
| "Topic" and its panel, after the trial series (`TaskCard.tsx:212–238`, R193, R194) | A third control in the row of "At most 2 actions, placed at the bottom of the card"; its panel opens in place, as "expanding and collapsing content sections" do | A third action | Move the choice into the full-screen progress, so that the row keeps two actions |
| The bar "Profile & progress", which turns the task card into the progress (`web/src/widget/CardFrame.tsx:111–116`, R97) | "No drill-ins, breadcrumbs, or multiple views"; full screen is for "data visualizations and dashboards" | "No deep navigation or multiple views within a card" | Open the progress full screen from the bar, so that the inline card stays one view and no second card is drawn, as R97 wants |
| The progress's sections and its switch of terms (T74.1, T74.3) | Expanding sections belong in the app; "segmented buttons" are the preferred control | "Cards should not contain multiple drill-ins, tabs, or deeper navigation" | Fine once the progress is full screen |
| The MathTrail mark in the header (R191, R203) | No rule | "Do not include your logo as part of the response. ChatGPT will always append your logo and app name before the widget is rendered" | Leave it out in ChatGPT; R191 left this to T71.1 |
| A palette of its own, not the host's (`internal/widget/tokens.css`) | "Use host tokens for all structural elements" | "Use system colors for text, icons, and spatial elements like dividers" | Keep: the design is the author's (R89, R156) |
| A turning icon while the task is written (`web/src/design/blocks.tsx:238–244`) | "Avoid spinners for inline content, because skeletons feel more native" | — | Keep; minor |
| Fonts | — | "Don't use custom fonts": the card uses the system's stack, `--font-sans` | Met |
| Light and dark themes, contrast, targets of 44 px | "All views must support both light and dark themes"; WCAG AA; 44 pt | WCAG AA | Met (`just web-layout`, `just web-preview`) |

**Recommendation:** a widget task before T71.4 that opens the progress full screen, moves the choice of a topic into it and leaves the mark out in ChatGPT, with everything else as it is and explained in the submission. The alternatives are everything as it is, answered at review, or a redesign to both guidelines in full, with the host's colours and at most two actions.

## 5. The Coach page, two steps from the card

**The path.** The card links a topic's page on the site (`web/src/widget/links.ts`: `https://mathtrail.app/{en,ru}/topics/<slug>/`). Every page of the site carries the menu, whose first entry is "Coach" (`web/src/site/frame.ts:40–50`, R202), and the Coach page frames the prototype of the author's coming product, with a free plan, a Plus plan at $10 a month and an "Upgrade" button (R198). The listing's website is the site's home page, under the same menu. The connector itself shows no advertisement, sells nothing and links to no plan.

**The rules**, all read 2026-10-05:

- Claude does not accept "Software that serves advertisements, sponsored content, paid product placements, or exists primarily as an advertising or promotional vehicle" ([Directory Policy](https://support.claude.com/en/articles/13145358-anthropic-software-directory-policy) 4.C), and rejects descriptions that "promote products and services" ([Checklist](https://claude.com/docs/connectors/building/review-criteria)). The form asks whether the connector handles "sponsored content" ([Submission](https://claude.com/docs/connectors/building/submission)).
- ChatGPT ([Plugin guidelines](https://developers.openai.com/plugins/plugin-guidelines)):
  - "Plugins must not serve advertisements and must not exist primarily as an advertising vehicle";
  - "Selling digital products or services—including subscriptions, digital content, tokens, or credits—is not allowed, whether offered directly or indirectly (for example, through freemium upsells)";
  - "Plugins must not display subscription plans, initiate new subscriptions, or promote upgrades", and may not "Link to a page that explicitly initiates the process to upgrade, subscribe, or complete a purchase";
  - subtitles and descriptions must not "advertise pricing, subscriptions, free trials, discounts, or promotions".

**The options:**

1. Keep it: the rules govern the plugin, and the plugin neither sells nor links a plan. The risk is a reviewer who follows the website and meets a paid plan in the menu's first entry; OpenAI's "indirectly" and "freemium upsells" are broad.
2. The topic pages the card opens show the menu without "Coach", and the home page keeps it. The listing's website is still the home page.
3. "Coach" leaves the menu and is linked from "About"; the page stays, with its price.
4. The Coach page without its price and its "Upgrade". R198 kept them on purpose, so this is listed only as an option.

**Recommendation: 3.** It takes the paid plan off every path from the card and off the listing's first page, and it keeps the page public. It reopens R202 (2026-10-05), which put "Coach" first in the menu, on evidence R202 did not have: the rules above.

## 6. Who submits

| | Claude | ChatGPT |
|---|---|---|
| Who may | "Plan: Pro, Max, Team, or Enterprise. Free accounts can't submit"; on Pro and Max from one's own account, on Team and Enterprise an Owner; "the listing belongs to the organization you submit from" ([Publish](https://claude.com/docs/directory/publish), read 2026-10-05) | "Complete individual or business verification … to publish under your name or a company name" ([Submission](https://developers.openai.com/plugins/deploy/submission)); "Publishing under an unverified individual or business name will result in rejection" ([Review requirements](https://developers.openai.com/plugins/deploy/app-review)); both read 2026-10-05 |
| What the listing shows | The company's name and website, asked in the form's Company step | The verified name, as the developer: it "is set automatically from your selected verified developer identity" ([Submission errors](https://developers.openai.com/plugins/deploy/submission-errors), read 2026-10-05) |
| What it takes | A paid plan on the account that submits and keeps the listing | Individual verification: an "original, physical government-issued ID" and, "if requested, a selfie" ([API organization verification](https://help.openai.com/en/articles/10910291-api-organization-verification), search summary: the page refused fetching, 2026-10-05); business verification: the company's documents; a project without EU data residency |

The project's contact address is the company's mailbox, `altedtech.info@gmail.com`; the domain is held by the author until T77, and Google Cloud is paid by the author until T76. Whether OpenAI's verification needs a paid API account was asked on OpenAI's forum on 5 October 2026 and is unanswered (§10).

**The options:** the author in person on both, the company on both, or one of each. **Recommendation:** one identity on both — the company, if it is a registered business that can pass OpenAI's business verification now; otherwise the author in person, knowing that ChatGPT shows the author's name. T77 left this to T71.6; it is asked now because the developer's name goes into T71.4's texts.

## 7. The requirements

Each row: an identifier, the requirement in the platform's words where they matter, the page and the day it was read, where MathTrail stands, and the status.

### 7.1 Claude

**The submission, in the order of the portal's steps**

| ID | Requirement | Source | MathTrail now | Status |
|---|---|---|---|---|
| CL-01 | "Plan: Pro, Max, Team, or Enterprise. Free accounts can't submit"; "there's no partner program to apply to first" | [Publish](https://claude.com/docs/directory/publish), read 2026-10-05 | — | Decision → О-65 |
| CL-02 | Connection: "the portal asks for an `https://` URL"; one URL for every user | [Submission](https://claude.com/docs/connectors/building/submission), read 2026-10-05 | `https://mcp.mathtrail.app/mcp` | Met |
| CL-03 | Tools: "grouped by whether their annotations declare them read-only or write"; tools "flagged for missing titles or annotations" are fixed first | [Submission](https://claude.com/docs/connectors/building/submission), read 2026-10-05 | Every tool has `title`, `readOnlyHint` and `destructiveHint` (`internal/transport/mcp/tool.go:101–130`); their values are CL-27 | Met; the values: Met in T71.2 (R215) |
| CL-04 | Listing: "server name up to 100 characters, one-liner up to 200 characters, description up to 2,000 characters, one to five categories, documentation URL, privacy policy URL, support contact, icon"; "The slug is permanent once published" | [Submission](https://claude.com/docs/connectors/building/submission), read 2026-10-05 | Nothing written yet | Gap → T71.4 (texts, icon, a slug proposed), T71.5 (the slug chosen) |
| CL-05 | Use cases: "what users need before they can connect" and "whether the connector reads data, writes data, or both" | [Submission](https://claude.com/docs/connectors/building/submission), read 2026-10-05 | A Claude account and a Google account; it reads and writes | Gap → T71.4 |
| CL-06 | Company: "Company name and website, plus a primary contact for review updates" | [Submission](https://claude.com/docs/connectors/building/submission), read 2026-10-05 | The contact is `altedtech.info@gmail.com` | Decision → О-65; Gap → T71.4 |
| CL-07 | Authentication: OAuth with dynamic client registration or a client ID metadata document | [Submission](https://claude.com/docs/connectors/building/submission), read 2026-10-05 | CIMD, with registration kept beside it (R03) | Met |
| CL-08 | Data handling: "Whether the underlying API is your own … and whether the connector handles personal health data or sponsored content" | [Submission](https://claude.com/docs/connectors/building/submission), read 2026-10-05 | MathTrail's own service, the parent's Drive as its storage; no health data; no sponsored content in the connector | Gap → T71.4; Decision → О-64 |
| CL-09 | Test and launch: "every link, credential, and step … including credentials for a fully populated account"; every tool run by the submitter | [Submission](https://claude.com/docs/connectors/building/submission), read 2026-10-05 | No reviewer account | Decision → О-62; Gap → T71.3 |
| CL-10 | Compliance: "Seven policy acknowledgments covering the directory guidelines, first-party API usage, financial transactions, AI media generation, prompt injection, conversation data collection, and public documentation. All seven are required." | [Submission](https://claude.com/docs/connectors/building/submission), read 2026-10-05 | Their words are shown only in the portal | Unknown → §10 |
| CL-11 | Screenshots of an MCP App: PNG, "at least 1000px", "3–5 images", "to the app response only, without the prompt in the image", the prompt given separately; "Video or GIF: not accepted" | [Submission](https://claude.com/docs/connectors/building/submission), read 2026-10-05 | The pictures of `docs/screens/` are the README's, not these | Gap → T71.4 (prompts), T71.5 (pictures); after О-63 |
| CL-12 | Allowed link URIs, optional: "Every origin and scheme you list must be owned by you" | [Submission](https://claude.com/docs/connectors/building/submission), read 2026-10-05 | The card opens `https://mathtrail.app/…` | Gap → T71.4 (`https://mathtrail.app`) |
| CL-13 | "Public documentation: required by your publish date. A blog post or help-center article is sufficient" | [Checklist](https://claude.com/docs/connectors/building/review-criteria), read 2026-10-05 | README's "Add it to your chat" (`README.md:58`), the home page's "connect" section, `docs/parents.md` | Met in part; Gap → T71.4 (which address) |

**The directory's policy and the checklist**

| ID | Requirement | Source | MathTrail now | Status |
|---|---|---|---|---|
| CL-14 | "Software must not violate or facilitate violation of our Usage Policy" (1.A) | [Directory Policy](https://support.claude.com/en/articles/13145358-anthropic-software-directory-policy), read 2026-10-05 | §2 | Decision → О-61 |
| CL-15 | "Software must only collect data from the user's context that is necessary to perform their function. Software must not collect extraneous conversation data, even for logging purposes" (1.D); no reading of Claude's memory, chat history or files (1.F) | [Directory Policy](https://support.claude.com/en/articles/13145358-anthropic-software-directory-policy), read 2026-10-05 | The tools take their own arguments and nothing else; the logs hold no text (R73) | Met |
| CL-16 | Descriptions "narrow, unambiguous" that "precisely match actual functionality" (2.A, 2.B); "Describe what the tool does, and don't tell Claude how to behave", and a description is rejected if it tells Claude "to behave in ways unrelated to the tool's function" | [Directory Policy](https://support.claude.com/en/articles/13145358-anthropic-software-directory-policy), [Checklist](https://claude.com/docs/connectors/building/review-criteria), read 2026-10-05 | Descriptions carry rules of behaviour tied to the tool, such as `submittask.go:158–165` | Gap → T71.2; Unknown → §10 |
| CL-17 | No calling or coercing Claude into calling other software, and no interfering with other tools (2.D, 2.E) | [Directory Policy](https://support.claude.com/en/articles/13145358-anthropic-software-directory-policy), read 2026-10-05 | The instructions name MathTrail's own tools alone | Met |
| CL-18 | "Instructional Software must not direct Claude to dynamically pull behavioral instructions from external sources for Claude to execute" (2.F); a description is rejected if it "Direct[s] Claude to pull behavioral instructions from external sources" | [Directory Policy](https://support.claude.com/en/articles/13145358-anthropic-software-directory-policy), [Checklist](https://claude.com/docs/connectors/building/review-criteria), read 2026-10-05 | `get_package` returns the guide the model writes a task by, with reference tasks and formats (`content/package.go:105–113`): the server's own text, embedded in it and public. It also carries the child's grade, interests and the parent's notes from the profile (`content/package.go:172–177`), delimited as information about the child, never as instructions (SPEC 4.1) | Unknown → §10; the letter asks |
| CL-19 | No "hidden, obfuscated, or encoded instructions" (2.G) | [Directory Policy](https://support.claude.com/en/articles/13145358-anthropic-software-directory-policy), read 2026-10-05 | The package is plain text; the sealed answer is data in the parent's file and never reaches the model | Met |
| CL-20 | A privacy policy "explaining data collection, usage, and retention" (3.A); "provide all applicable third-party privacy policies" | [Directory Policy](https://support.claude.com/en/articles/13145358-anthropic-software-directory-policy), [Directory Terms](https://support.claude.com/en/articles/13145338-anthropic-software-directory-terms), read 2026-10-05 | `https://mathtrail.app/en/privacy/`; a revision of it was under way on 2026-10-05, so T71.4 checks the version published when it runs | Met in part; Gap → T71.4 (line by line) |
| CL-21 | "verified contact information and support channels for users with product or security concerns" (3.B); "a mechanism for receiving reports of security vulnerabilities" | [Directory Policy](https://support.claude.com/en/articles/13145358-anthropic-software-directory-policy), [Directory Terms](https://support.claude.com/en/articles/13145338-anthropic-software-directory-terms), read 2026-10-05 | `altedtech.info@gmail.com` in the policy and the terms; GitHub's private reporting (`SECURITY.md`) | Met in part; Gap → T71.4 (a support page) |
| CL-22 | "document how their Software works, its intended purpose, and how users can troubleshoot issues" (3.C) | [Directory Policy](https://support.claude.com/en/articles/13145358-anthropic-software-directory-policy), read 2026-10-05 | README, `docs/parents.md`; nothing on troubleshooting | Gap → T71.4 |
| CL-23 | "a standard testing account with sample data" (3.D) | [Directory Policy](https://support.claude.com/en/articles/13145358-anthropic-software-directory-policy), read 2026-10-05 | — | Decision → О-62; Gap → T71.3 |
| CL-24 | "at least three working examples of prompts or use cases" (3.E) | [Directory Policy](https://support.claude.com/en/articles/13145358-anthropic-software-directory-policy), read 2026-10-05 | — | Gap → T71.4 |
| CL-25 | "verify that they own or control any API endpoint, domain, or user interface their Software connects to" (3.F); "Your server must call your own first-party APIs, or APIs you legitimately proxy. The MCP server domain should match your service" | [Directory Policy](https://support.claude.com/en/articles/13145358-anthropic-software-directory-policy), [Checklist](https://claude.com/docs/connectors/building/review-criteria), read 2026-10-05 | `mcp.mathtrail.app` under `mathtrail.app`; Drive is reached as the parent's OAuth client, on the parent's grant | Met |
| CL-26 | "follow design guidelines Anthropic publishes applicable to Software" (3.H) | [Directory Policy](https://support.claude.com/en/articles/13145358-anthropic-software-directory-policy), read 2026-10-05 | §4 | Decision → О-63 |
| CL-27 | "Every tool must include a `title` and the applicable hint: `readOnlyHint: true` for read-only tools, and `destructiveHint: true` for tools that modify or delete data"; "Read-only tools can run without per-call confirmation, and destructive tools always prompt" (5.E); a tool for both reading and writing "is rejected" | [Directory Policy](https://support.claude.com/en/articles/13145358-anthropic-software-directory-policy), [Checklist](https://claude.com/docs/connectors/building/review-criteria), read 2026-10-05 | Five tools marked read-only may write a repaired file (R119); `save_profile` and `edit_profile` overwrite the file; `submit_answer` changes the ratings; every tool has `destructiveHint: false` (SPEC 7.1) | Met in T71.2, by the documentation with no live check (R214, R215); what Claude does with a destructive tool a card calls: Unknown → §10 |
| CL-28 | No "advertisements, sponsored content, paid product placements" (4.C) | [Directory Policy](https://support.claude.com/en/articles/13145358-anthropic-software-directory-policy), read 2026-10-05 | None in the connector; the site's Coach page, §5 | Met for the connector; Decision → О-64 |
| CL-29 | "gracefully handle errors and provide helpful feedback rather than generic error messages" (5.A); inputs validated, with "actionable error messages" | [Directory Policy](https://support.claude.com/en/articles/13145358-anthropic-software-directory-policy), [Checklist](https://claude.com/docs/connectors/building/review-criteria), read 2026-10-05 | One sentence for each failure (`internal/transport/mcp/failure.go:51–202`); a refusal names every reason | Met |
| CL-30 | "MCP servers must be frugal with their use of tokens" (5.B); "Keep responses reasonably sized" | [Directory Policy](https://support.claude.com/en/articles/13145358-anthropic-software-directory-policy), [Checklist](https://claude.com/docs/connectors/building/review-criteria), read 2026-10-05 | A package is 24.2 KB on average and 38.1 KB at most, at every limit of the profile, under a budget of 64 KB (SPEC 4.1); a task adds about 25 KB to the conversation (`docs/live/09-load.md:238`) | Met |
| CL-31 | Tool names of at most 64 characters (5.C); OAuth 2.0 "with certificates from recognized authorities" (5.D); Streamable HTTP (5.F) | [Directory Policy](https://support.claude.com/en/articles/13145358-anthropic-software-directory-policy), read 2026-10-05 | The longest name has 13; Google's managed certificate; Streamable HTTP | Met |

**The service**

| ID | Requirement | Source | MathTrail now | Status |
|---|---|---|---|---|
| CL-32 | A `401` whose `WWW-Authenticate` points at the protected resource's metadata; its `resource` equal to the URL; the first of `authorization_servers` used | [Authentication](https://claude.com/docs/connectors/building/authentication), read 2026-10-05 | SPEC 9.4; signed in live since T53.2 | Met |
| CL-33 | A client ID metadata document is used only when the metadata has `client_id_metadata_document_supported: true` and `none`; S256 advertised; the callback `https://claude.ai/api/mcp/auth_callback` | [Authentication](https://claude.com/docs/connectors/building/authentication), read 2026-10-05 | All three (`internal/transport/oauth/metadata.go:58–104`); the callback comes in Claude's own document | Met |
| CL-34 | Claude Code: "accept a loopback redirect on any port", for `localhost` and `127.0.0.1` | [Authentication](https://claude.com/docs/connectors/building/authentication), read 2026-10-05 | A loopback redirect matches with its port (SPEC 9.3, remark 49), so Claude Code cannot sign in | Met in T71.2 (R218) |
| CL-35 | Refresh "reactively on a `401` response, and proactively up to five minutes before the stored expiry"; `invalid_grant`; refresh tokens rotated; `/token` takes a form; 10 seconds for discovery, registration and token, 30 for a refresh | [Authentication](https://claude.com/docs/connectors/building/authentication), read 2026-10-05 | R113: `invalid_grant`, a new refresh token on each use, the old one good to its own end (SPEC 9.3); the token endpoint may renew at Google first (R117) | Met |
| CL-36 | "Anthropic's outbound traffic to your server originates from `160.79.104.0/21`" | [Authentication](https://claude.com/docs/connectors/building/authentication), read 2026-10-05 | The sign-in's endpoints are held to 20 a minute an address (R121) | Met in T71.2: the pace stays, with the sign to raise it (R219) |
| CL-37 | A result of "\~150,000 characters" at most on claude.ai and Desktop, 25,000 tokens in Claude Code; "240 seconds per tool call" | [Build an MCP server](https://claude.com/docs/connectors/building/index), read 2026-10-05 | A package is 38.1 KB at most (SPEC 4.1), far under 150,000 characters; Claude Code's limit is counted in tokens and was not measured, and Claude Code cannot sign in today (CL-34) | Met on claude.ai; Claude Code: Unknown → §10 |
| CL-38 | `_meta.ui.domain` is optional and must be exactly `{hash}.claudemcpcontent.com`; "when validation fails it shows an `Invalid ui.domain format` or `ui.domain mismatch` error instead of rendering the app" | [Get started with MCP Apps](https://claude.com/docs/connectors/building/mcp-apps/getting-started), [Troubleshoot MCP Apps](https://claude.com/docs/connectors/building/mcp-apps/troubleshooting), read 2026-10-05 | Not set (SPEC 8.1) | Met; see §7.3 |
| CL-39 | The CSP through `_meta.ui.csp`; "`frameDomains` … is restricted in Claude" | [Design guidelines](https://claude.com/docs/connectors/building/mcp-apps/design-guidelines), read 2026-10-05 | All four lists empty | Met |
| CL-40 | The inline card: "At most 2 actions", "At most 4-5 data points", "No drill-ins, breadcrumbs, or multiple views", "No menus or popovers" | [Design guidelines](https://claude.com/docs/connectors/building/mcp-apps/design-guidelines), read 2026-10-05 | §4 | Decision → О-63 |
| CL-41 | `ui/open-link` asks the person first, unless the listing allows the destination and the request follows "a real user gesture" | [External links](https://claude.com/docs/connectors/building/mcp-apps/external-links), read 2026-10-05 | Links open on a press (`web/src/widget/bridge.ts:211–219`) | Met; the list: CL-12 |
| CL-42 | A card's `ui/message`: no page says a directory connector's message skips the confirmation | [MCP Apps pages](https://claude.com/docs/llms.txt), read 2026-10-05 | Claude puts the card's message into the input under a warning (finding 1 of `docs/live/05-login-drive.md`) | Unknown → §10 |

**Review and after publication**

| ID | Requirement | Source | MathTrail now | Status |
|---|---|---|---|---|
| CL-43 | "Anthropic scans your submission automatically … and, by default, lists it as a Community connector"; Verified is an escalation nobody applies for; "Review time isn't fixed" | [Submission](https://claude.com/docs/connectors/building/submission), [Publish](https://claude.com/docs/directory/publish), read 2026-10-05 | — | N/A |
| CL-44 | An approved listing goes live once its owner selects Publish; "If you edit the listing before you publish, it returns to draft for another review" | [Submission status](https://claude.com/docs/directory/submission-status), read 2026-10-05 | — | N/A; T71.5 |
| CL-45 | "You don't resubmit to Anthropic to change a listed MCP server's tools"; "The tool names shown on your directory listing are part of the listing details, so update them with a listing edit, which a reviewer approves" | [After publishing](https://claude.com/docs/connectors/building/after-publishing), read 2026-10-05 | — | N/A (§12) |
| CL-46 | Listing edits wait for a reviewer; a new display name "requires re-review"; "The URL slug is locked"; a connector is delisted by email; a change of the server's address is not described | [Managing your listing](https://claude.com/docs/connectors/building/managing-your-listing), [After publishing](https://claude.com/docs/connectors/building/after-publishing), read 2026-10-05 | — | N/A; the address: Unknown → §10 |
| CL-47 | The health badge: "Request errors include tool calls rejected for authentication problems and exclude errors a tool returns in its own result"; Healthy at 2 % or less, "Worth a look" above 2 %, "Degraded" above 5 %; the error rate counts results with `isError: true` as well | [Managing your listing](https://claude.com/docs/connectors/building/managing-your-listing), read 2026-10-05 | §8.2 | Met by design; the rates: Unknown → §10 |
| CL-48 | Anthropic "may remove or refuse to display any Software … at any time for any reason"; an indemnity; a licence to show the name and the logos | [Directory Terms](https://support.claude.com/en/articles/13145338-anthropic-software-directory-terms), read 2026-10-05 | — | N/A |
| CL-49 | Users of 18 and over; no account "available to anyone else"; accounts disabled on "indicators of minor activity" | [Consumer Terms](https://www.anthropic.com/legal/consumer-terms), [Age assurance](https://support.claude.com/en/articles/15171100-age-assurance-on-claude), read 2026-10-05 | §2 | Decision → О-61 |

### 7.2 ChatGPT

**The submission and the package**

| ID | Requirement | Source | MathTrail now | Status |
|---|---|---|---|---|
| GPT-01 | "All plugin submissions must come from verified individuals or organizations"; "Publishing under an unverified individual or business name will result in rejection"; the directory shows "the name associated with this identity" | [Plugin guidelines](https://developers.openai.com/plugins/plugin-guidelines), [Review requirements](https://developers.openai.com/plugins/deploy/app-review), [Submission](https://developers.openai.com/plugins/deploy/submission), read 2026-10-05 | — | Decision → О-65; Gap → T71.6 |
| GPT-02 | "Organization owners can submit; other members need Apps Management Write"; "projects with EU data residency cannot submit plugins with MCP servers for review" | [Submission](https://developers.openai.com/plugins/deploy/submission), [Review requirements](https://developers.openai.com/plugins/deploy/app-review), read 2026-10-05 | — | Gap → T71.6 |
| GPT-03 | A ZIP with `plugin.json` and `mcp.json`; "only one MCP server can be connected per plugin"; no credentials in it; at most 100 MB | [Submission](https://developers.openai.com/plugins/deploy/submission), [Submission errors](https://developers.openai.com/plugins/deploy/submission-errors), read 2026-10-05 | — | Gap → T71.4 |
| GPT-04 | Display name and subtitle of at most 30 characters each; a long description of at most 4,000; a developer name of at most 80; one category of thirteen, "Education & Research" among them; at most 20 capabilities of 120 characters; at most three starter prompts of 128; the package's `description` at most 1,024 at the final submission, where the submission page says 4,000 | [Submission](https://developers.openai.com/plugins/deploy/submission), [Submission errors](https://developers.openai.com/plugins/deploy/submission-errors), read 2026-10-05 | — | Gap → T71.4 |
| GPT-05 | A website, a "customer support page", a privacy policy and terms, each over HTTPS; "Public URLs must be accessible and identify the same publisher as the submission" | [Submission](https://developers.openai.com/plugins/deploy/submission), read 2026-10-05 | The home page, `/en/privacy/` and `/en/terms/` on `mathtrail.app`; no support page | Gap → T71.4 |
| GPT-06 | An icon, "square and at least 48 by 48 pixels", PNG, JPEG, WebP or SVG of at most 5 MiB; a dark variant optional | [Submission](https://developers.openai.com/plugins/deploy/submission), read 2026-10-05 | T61a's logo, `site/assets/favicon.svg` | Gap → T71.4 |
| GPT-07 | No "MCP", "MCP Server" or "Plugin" after the name; subtitles and descriptions do not "advertise pricing, subscriptions, free trials, discounts, or promotions" | [Plugin guidelines](https://developers.openai.com/plugins/plugin-guidelines), read 2026-10-05 | The README sells MathTrail as free | Gap → T71.4 (no price as a selling point) |
| GPT-08 | The domain's token, "as plain text", at `/.well-known/openai-apps-challenge` "on the MCP hostname or an eligible parent domain" | [Submission](https://developers.openai.com/plugins/deploy/submission), [Submission errors](https://developers.openai.com/plugins/deploy/submission-errors), read 2026-10-05 | No such address. On the site, GitHub Pages would serve a file with no extension as `application/octet-stream`; the pages' archive keeps hidden folders (`.github/workflows/pages.yml:139–141`) | Met in T71.2: the service serves it from `MATHTRAIL_OPENAI_CHALLENGE`, which T71.6 sets (R216); whether the check takes it: Unknown → §10 |
| GPT-09 | The reviewer's account "should work immediately without MFA approval, email or SMS codes, magic links, or private-network access"; credentials tested "outside any company networks"; "Keep the test account and sample data available for subsequent reviews" | [Submission](https://developers.openai.com/plugins/deploy/submission), [Review requirements](https://developers.openai.com/plugins/deploy/app-review), read 2026-10-05 | — | Decision → О-62; Gap → T71.3 |
| GPT-10 | "Exactly five positive test cases, three negative test cases, and release notes"; a positive case names the prompt, the tools and the expected result; "A demo-recording URL that shows the main use cases and tools across supported platforms" | [Submission errors](https://developers.openai.com/plugins/deploy/submission-errors), [Submission](https://developers.openai.com/plugins/deploy/submission), read 2026-10-05 | — | Gap → T71.4 (cases), T71.6 (the recording) |
| GPT-11 | Screenshots, optional with UI: "exactly 706 pixels wide and 400–860 pixels tall", one for each starter prompt; and yet "Screenshots are no longer shown in the Directory, instead provide example Prompts" | [Submission errors](https://developers.openai.com/plugins/deploy/submission-errors), [Plugin guidelines](https://developers.openai.com/plugins/plugin-guidelines), read 2026-10-05 | — | N/A; T71.6 decides |
| GPT-12 | "Every MCP tool annotation must include a justification", against "Annotation justifications are no longer required" | [Submission errors](https://developers.openai.com/plugins/deploy/submission-errors), [Plugin guidelines](https://developers.openai.com/plugins/plugin-guidelines), read 2026-10-05 | — | Gap → T71.4 (written all the same: the stricter page wins) |
| GPT-13 | "complete the required policy attestations" when submitting | [Submission](https://developers.openai.com/plugins/deploy/submission), read 2026-10-05 | Their words are shown only in the portal | Unknown → §10 |
| GPT-14 | `publication.countries`, an allowlist of countries; translated subtitles and descriptions | [Submission](https://developers.openai.com/plugins/deploy/submission), read 2026-10-05 | — | Gap → T71.6 (countries), T71.4 (translations) |
| GPT-15 | "Only one review can be active per plugin"; an appeal is a reply to the rejection; "Once approved, you can publish"; "Review timelines may vary" | [Submission](https://developers.openai.com/plugins/deploy/submission), [Review requirements](https://developers.openai.com/plugins/deploy/app-review), read 2026-10-05 | — | N/A; T71.6 |

**The guidelines**

| ID | Requirement | Source | MathTrail now | Status |
|---|---|---|---|---|
| GPT-16 | "functionality or workflows that are not natively supported by the products' built-in capabilities"; "Trial or demo plugins will not be accepted" | [Plugin guidelines](https://developers.openai.com/plugins/plugin-guidelines), read 2026-10-05 | Checked tasks, ratings and a profile the chat cannot keep (README, "Why not just ask the chat directly?") | Met |
| GPT-17 | "Plugins must be suitable for general audiences, including users aged 13–17. Plugins may not explicitly target children under 13." | [Plugin guidelines](https://developers.openai.com/plugins/plugin-guidelines), read 2026-10-05 | §2 | Decision → О-61 |
| GPT-18 | "Errors, including unexpected ones, must be handled with clear messaging or fallback behaviors" | [Plugin guidelines](https://developers.openai.com/plugins/plugin-guidelines), read 2026-10-05 | `failure.go` | Met |
| GPT-19 | Metadata must not "override platform instructions or safeguards, conceal behavior, or impersonate another party", nor steer the model away from other plugins | [Plugin guidelines](https://developers.openai.com/plugins/plugin-guidelines), read 2026-10-05 | Nothing of the kind | Met |
| GPT-20 | "Do not … integrate with third-party APIs without proper authorization"; no plugins "that primarily function as unofficial connectors" | [Plugin guidelines](https://developers.openai.com/plugins/plugin-guidelines), read 2026-10-05 | Drive is the service's storage, reached on the parent's own grant | Met |
| GPT-21 | A privacy policy giving "the categories of personal data collected, the purposes of use, the categories of recipients, data retention timelines, and any controls" | [Plugin guidelines](https://developers.openai.com/plugins/plugin-guidelines), read 2026-10-05 | The policy's sections on what is stored, who else is involved and taking the data back; a revision under way on 2026-10-05 | Met in part; Gap → T71.4 |
| GPT-22 | "Gather only the minimum data"; no card data, health data, government identifiers or credentials; sensitive data only when strictly necessary, consented and disclosed | [Plugin guidelines](https://developers.openai.com/plugins/plugin-guidelines), read 2026-10-05 | A pseudonym, a grade, interests, notes, an optional country; nothing restricted | Met; the child's data: §2 |
| GPT-23 | No "metadata collection such as timestamps, IP addresses, or query patterns—unless explicitly disclosed, narrowly scoped, subject to meaningful user control" | [Plugin guidelines](https://developers.openai.com/plugins/plugin-guidelines), read 2026-10-05 | The country is read from the sign-in's address and disclosed ("The country you sign in from"), but the family has no way to turn it off (R185) | Met in T71.2: a tick in the profile, and the policy says how (R220) |
| GPT-24 | No digital goods "directly or indirectly (for example, through freemium upsells)"; plugins "must not display subscription plans, initiate new subscriptions, or promote upgrades", nor "Link to a page that explicitly initiates the process to upgrade" | [Plugin guidelines](https://developers.openai.com/plugins/plugin-guidelines), read 2026-10-05 | The plugin sells nothing; the Coach page, §5 | Decision → О-64 |
| GPT-25 | "Plugins must not serve advertisements and must not exist primarily as an advertising vehicle" | [Plugin guidelines](https://developers.openai.com/plugins/plugin-guidelines), read 2026-10-05 | None | Met; О-64 |
| GPT-26 | "You must provide customer support contact details where end users can reach you for help" | [Plugin guidelines](https://developers.openai.com/plugins/plugin-guidelines), read 2026-10-05 | An address in the policy | Gap → T71.4 (with GPT-05) |
| GPT-27 | Tool names in "plain language that directly reflects the action, ideally as a verb", unique | [Plugin guidelines](https://developers.openai.com/plugins/plugin-guidelines), read 2026-10-05 | `get_profile`, `next_task`, `submit_answer`… | Met |
| GPT-28 | A description explains "purpose and behavior, when to use it, and any relevant limitations or side effects", and does not "recommend overly broad triggering" | [Plugin guidelines](https://developers.openai.com/plugins/plugin-guidelines), read 2026-10-05 | No description states the repair a read may make (R119) | Gap → T71.2 |
| GPT-29 | "Each public tool must operate independently. Do not instruct the model to invoke another plugin or connector" | [Plugin guidelines](https://developers.openai.com/plugins/plugin-guidelines), read 2026-10-05 | MathTrail's tools call for its own | Met |
| GPT-30 | `readOnlyHint`, `destructiveHint` and `openWorldHint` as explicit booleans; `readOnlyHint` false if a tool "can create/update/delete anything"; `destructiveHint` true for "deleting, overwriting … even in only select modes, through default parameters, or through indirect side effects", false "only for additive writes"; `openWorldHint` false for "a bounded private account" | [Plugin guidelines](https://developers.openai.com/plugins/plugin-guidelines), [Review requirements](https://developers.openai.com/plugins/deploy/app-review), [Reference](https://developers.openai.com/plugins/reference), read 2026-10-05 | All three explicit; `openWorldHint: false` is right; the other two as CL-27, and every day's first write may delete the earliest of a hundred kept revisions (R119, `internal/store/drive/history.go:156–167`) | Met in T71.2 (R214, R215) |
| GPT-31 | Minimal inputs; no conversation history; "Avoid requesting raw location fields (for example, city or coordinates)" | [Plugin guidelines](https://developers.openai.com/plugins/plugin-guidelines), read 2026-10-05 | `country` and `region` are optional ISO codes (R185) | Met |
| GPT-32 | "Side effects should never be hidden or implicit"; tools "safe to retry" or saying they are not | [Plugin guidelines](https://developers.openai.com/plugins/plugin-guidelines), read 2026-10-05 | R119's repair on a read; `submit_task` says each call spends an attempt | Met in T71.2: no read mends, and every description says what its tool changes in the file (R214, R215) |
| GPT-33 | No "session IDs, trace IDs, request IDs, timestamps, or logging metadata—unless they are strictly required to fulfill the user's query" | [Plugin guidelines](https://developers.openai.com/plugins/plugin-guidelines), [Review requirements](https://developers.openai.com/plugins/deploy/app-review), read 2026-10-05 | §8: `answered_at`, `age_seconds` with the seconds `next_task`'s text gives, and `instructions_version` are not needed; `tsk_…` and `req_…` are, since the model passes them back, as "stable identifiers in structured results so later tools can refer to the same records" ([Build an MCP server](https://developers.openai.com/plugins/build/mcp-server)) | Met in T71.2 (R217) |
| GPT-34 | The server "must not pull, reconstruct, or infer the full chat log" | [Plugin guidelines](https://developers.openai.com/plugins/plugin-guidelines), read 2026-10-05 | Nothing of the kind | Met |
| GPT-35 | An authentication flow "transparent and explicit", its permissions "limited to what is necessary" | [Plugin guidelines](https://developers.openai.com/plugins/plugin-guidelines), read 2026-10-05 | The consent screen; `drive.file` alone | Met |
| GPT-36 | A justification for every embedded frame | [Plugin guidelines](https://developers.openai.com/plugins/plugin-guidelines), read 2026-10-05 | No frames | N/A |
| GPT-37 | "Plugins must function reliably in ChatGPT on both desktop and mobile, including any UI components"; test cases pass "on the supported ChatGPT and Codex surfaces" | [Plugin guidelines](https://developers.openai.com/plugins/plugin-guidelines), [Review requirements](https://developers.openai.com/plugins/deploy/app-review), read 2026-10-05 | ChatGPT is untested (T63); widgets on phones are unconfirmed (PRODUCT 9.4); Codex was never tried | Unknown → §10 |
| GPT-38 | Server instructions: "Keep the most important details in the first 512 characters. Do not repeat every tool description or try to change the model's personality" | [Build an MCP server](https://developers.openai.com/plugins/build/mcp-server), read 2026-10-05 | 9,327 characters; the first 512 hold the role and the first rule | Unknown → §10 |

**The service**

| ID | Requirement | Source | MathTrail now | Status |
|---|---|---|---|---|
| GPT-39 | Discovery by the protected resource's metadata or a `401`; the server's metadata with `issuer`, the endpoints, `client_id_metadata_document_supported`, `none` and S256 | [Authentication](https://developers.openai.com/plugins/build/auth), read 2026-10-05 | `metadata.go:35–104` | Met |
| GPT-40 | "Return `iss` in every successful and error authorization response. Its value must exactly match the metadata `issuer`"; ChatGPT then uses `https://chatgpt.com/connector_platform_oauth_redirect` | [Authentication](https://developers.openai.com/plugins/build/auth), read 2026-10-05 | `iss` on every redirect to a client (`internal/transport/oauth/flow.go:50–57`), one issuer value | Met |
| GPT-41 | ChatGPT's document `https://chatgpt.com/oauth/client.json` gives `"token_endpoint_auth_method": "private_key_jwt"` as a preference and `"token_endpoint_auth_methods_supported": ["none", "private_key_jwt"]`; ChatGPT uses a method both sides list | [Authentication](https://developers.openai.com/plugins/build/auth), the document itself, read 2026-10-05 | Our metadata lists `none`, and the document reader refuses only secrets (`internal/infra/cimd/document.go:34–64`) | Met on reading; live in T63 |
| GPT-42 | A `UserInfo` endpoint with `email` and `email_verified` — "required for workspace domain restrictions" of ChatGPT Enterprise | [Authentication](https://developers.openai.com/plugins/build/auth), read 2026-10-05 | None; Google is asked for `openid` and `drive.file` | N/A |
| GPT-43 | "Keep the registered OAuth client and any client secret valid while the MCP server connection is in use" | [Authentication](https://developers.openai.com/plugins/build/auth), read 2026-10-05 | CIMD needs no registration; a registered client's identifier is sealed and lasts while its key does | Met |
| GPT-44 | The outbound addresses of "ChatGPT integrations", `https://openai.com/chatgpt-connectors.json`: 278 prefixes, some 36,000 addresses, created 2026-09-22; "The ranges can change" | [IP egress ranges](https://developers.openai.com/api/docs/guides/ip-addresses), read 2026-10-05 | The pace of §8.1 | Met in T71.2 (R219) |
| GPT-45 | `_meta.ui.domain`: "required when submitting a plugin with UI; must be unique per plugin"; `_meta["openai/widgetDomain"]` is an "OpenAI-specific compatibility alias" | [Reference](https://developers.openai.com/plugins/reference), read 2026-10-05 | Not set | Met in T71.2 by `openai/widgetDomain`, `ui.domain` left unset (R216); whether ChatGPT takes the alias: Unknown → §10 |
| GPT-46 | A tool links its UI by `_meta.ui.resourceUri` | [Reference](https://developers.openai.com/plugins/reference), read 2026-10-05 | `get_progress` and `next_task` | Met |
| GPT-47 | The CSP's `connectDomains`, `resourceDomains`, `frameDomains`; `_meta["openai/widgetCSP"].redirect_domains` "is still required for trusted `openExternal` destinations" | [Reference](https://developers.openai.com/plugins/reference), read 2026-10-05 | All empty; no `redirect_domains` | Met in T71.2: `redirect_domains` is the site's origin (R216) |
| GPT-48 | `securitySchemes` on the tool, with `_meta["securitySchemes"]` as a "Back-compat mirror for clients that only read `_meta`" | [Reference](https://developers.openai.com/plugins/reference), read 2026-10-05 | `_meta` alone, since go-sdk v1.8.0's `Tool` has no such field (SPEC remark 20) | Left to T63, which sees whether ChatGPT needs the field (R216) |
| GPT-49 | "Treat the resource URI as a cache key"; "ChatGPT may continue serving cached resource contents for up to one hour"; "Keep existing input schemas and each published UI resource URI working" | [Add UI](https://developers.openai.com/plugins/build/chatgpt-ui), [Review requirements](https://developers.openai.com/plugins/deploy/app-review), read 2026-10-05 | One fixed address (SPEC 8.1, remark 38) | Decision → О-66 |

**After publication, and the terms**

| ID | Requirement | Source | MathTrail now | Status |
|---|---|---|---|---|
| GPT-50 | Scans every day; "Deleted tools: Removed … as soon as a scan detects the deletion"; "New tools: Made available after they pass automated checks"; "Changed tools: The previous definition stays live until the updated definition passes automated checks"; "Keep your server compatible with the live definition while an update is held" | [Review requirements](https://developers.openai.com/plugins/deploy/app-review), [Submission](https://developers.openai.com/plugins/deploy/submission), read 2026-10-05 | — | N/A (§12) |
| GPT-51 | A tool's `_meta`, its security schemes, its UI resource and its CSP, and the server's `instructions`, are reviewed with the tools | [Review requirements](https://developers.openai.com/plugins/deploy/app-review), read 2026-10-05 | — | N/A (§12) |
| GPT-52 | "changes to plugin metadata or skills still require a new ZIP", and a new review | [Submission](https://developers.openai.com/plugins/deploy/submission), read 2026-10-05 | — | N/A |
| GPT-53 | "The MCP server origin (`scheme`, `hostname`, or `port`) can't change between versions"; the path can, in a new version; the submission page says "To change an existing MCP server's URL, contact support" | [Review requirements](https://developers.openai.com/plugins/deploy/app-review), [Submission](https://developers.openai.com/plugins/deploy/submission), read 2026-10-05 | The address stays | N/A |
| GPT-54 | "Plugins appear on the directory's main pages only if OpenAI selects them for enhanced distribution"; a press release goes through `press@openai.com` first | [Review requirements](https://developers.openai.com/plugins/deploy/app-review), read 2026-10-05 | — | N/A |
| GPT-55 | Users from 13, under 18 with a parent's permission; no account "available to anyone else" | [Terms of Use](https://openai.com/policies/row-terms-of-use/), search summary: the page refused fetching, 2026-10-05 | §2 | Decision → О-61 |

### 7.3 Where one value cannot serve both hosts

- **The widget's sandbox domain.** Claude takes `{hash}.claudemcpcontent.com` or nothing, and draws an error instead of a card on any other value (CL-38); ChatGPT wants a domain unique to the plugin (GPT-45). Two ways serve both. Either `ui.domain` stays unset, as Claude's default, and `openai/widgetDomain` carries ChatGPT's value, though whether ChatGPT's check takes the alias is learnt only at upload (§10). Or the service answers each host with its own `ui.domain`, telling the hosts apart by the client's name as it already does to answer a lost sign-in (`internal/transport/mcp/labels.go:48–57`, R118). T71.2 chooses.
- **What "destructive" means.** Claude: tools that "modify or delete data", and "destructive tools always prompt" (CL-27). ChatGPT: anything that overwrites or may not be undone, "even … through indirect side effects", and false "only for additive writes" (GPT-30). One value per tool has to satisfy both, and Claude's reading asks the person before every write.
- **Pictures.** Claude: three to five PNGs at least 1,000 px wide, cropped to the card, without the prompt (CL-11). ChatGPT: optional, exactly 706 px wide, one for each starter prompt, and "no longer shown in the Directory" (GPT-11).
- **The mark.** Claude has no rule against it in the card; ChatGPT asks it out (§4).
- **The outbound addresses.** One /21 for Claude; a list of 278 prefixes for ChatGPT that changes (CL-36, GPT-44).
- **The listing's texts.** A name of 100 characters against 30, a line of 200 against a subtitle of 30, a description of 2,000 against 4,000 (CL-04, GPT-04).

## 8. The gaps

The audit of 2026-09-30, item by item, then the gaps found since. "Open" means a task or a question still owns it; T71.5 and T71.6 start only when every row is closed or accepted as it is.

| # | Gap | Verdict | Rows | Owner | State |
|---|---|---|---|---|---|
| 1 | `serverInfo` has no `icons`, `websiteUrl` or `description` (`internal/transport/mcp/server.go:103`) | Confirmed in the code, and asked for by neither directory: Claude shows the listing's icon (a custom connector gets its domain's favicon, R191), ChatGPT the package's | CL-04, GPT-06 | T71.2, optional | Closed in T71.2 (R216) |
| 2 | Hints against behaviour: read-only tools that may write (R119), `save_profile` overwriting with `destructiveHint: false`, the tools that add to history and ratings | Confirmed, and wider: both definitions count an overwrite, and ChatGPT counts "indirect side effects", such as the earliest kept revision a day's first write may delete | CL-27, GPT-30, GPT-32 | T71.2 | Closed in T71.2, by the documentation (R214, R215) |
| 3 | `_meta.ui.domain`; a version in the widget's address (remark 38); `securitySchemes` in `_meta` alone (remark 20) | Confirmed. Changed: Claude refuses a value not its own, so the value is per host; ChatGPT's cache of an hour answers remark 38 | CL-38, GPT-45, GPT-48, GPT-49 | T71.2; О-66; T63 | Open: the domain closed in T71.2 by `openai/widgetDomain` (R216); the version of the address waits on О-66, and `securitySchemes` on the tool on T63 |
| 4 | The ChatGPT sign-in: `iss`, and `UserInfo` with an email | Dropped: `iss` is on every redirect, and `UserInfo` serves Enterprise domain restrictions alone | GPT-40, GPT-42 | — | Closed |
| 5 | Timestamps and internal identifiers in results | Confirmed in part: `last_answer.answered_at`, `recent[].answered_at` (`internal/transport/mcp/payload.go:53–73`, `progress.go:108–113`), `age_seconds` (`nexttask.go:85`) with the seconds in `next_task`'s text (`nexttask.go:217`), and `instructions_version` (`content/package.go:112`) go; `tsk_…`, `req_…` and the Drive links stay | GPT-33 | T71.2 | Closed in T71.2 (R217) |
| 6 | The card against the design rules | Confirmed, and Claude makes its guidelines binding | CL-26, CL-40, §4 | О-63 | Open |
| 7 | The pace by address against the hosts' servers (remark 60) | Confirmed; counted in §8.1 | CL-36, GPT-44 | T71.2 | Closed in T71.2 (R219) |
| 8 | Claude's health indicator counting a refused `submit_task` | Dropped: a refusal is an ordinary result, and the badge leaves out errors a tool returns in its own result. New: the badge counts sign-ins refused (§8.2) | CL-47 | §10 | Closed |
| 9 | Claude's warning on the card's `ui/message` (finding 1 of report 05) | Unknown: no page describes it | CL-42 | §10; T71.5 sees it | Open |
| 10 | The form's questions: a "company" for one developer, a support address, public documentation, ChatGPT's countries | Confirmed | CL-06, CL-13, GPT-05, GPT-14 | О-65; T71.4; T71.6 | Open |
| 11 | The card's links to topic pages (T74.6) against the rules on links | Fine with two settings: Claude's allowed link URIs and ChatGPT's `redirect_domains` | CL-12, CL-41, GPT-47 | T71.4; T71.2 | Open: ChatGPT's `redirect_domains` set in T71.2 (R216); Claude's allowed link URIs are T71.4's |
| 12 | The Coach page in the menu (T75.14, R198) against the rules on advertising | Confirmed as a risk | CL-28, GPT-24 | О-64 | Open |
| 13 | Claude Code's loopback redirect on a port of its own | New | CL-34 | T71.2: accept any port for a client that names itself by a document, or say Claude Code is not supported | Closed in T71.2: any port for a client of its own document (R218) |
| 14 | The package against "behavioral instructions from external sources" | New | CL-18 | The letter to Anthropic; §10 | Open |
| 15 | Descriptions that tell Claude how to behave | New | CL-16 | The task that О-61 brings: the descriptions move with the instructions; T71.2 added what each tool changes in the file | Open |
| 16 | No support page | New | CL-21, GPT-05, GPT-26 | T71.4 | Open |
| 17 | ChatGPT's domain token may sit on the site, a parent domain | New option | GPT-08 | T71.2 | Closed in T71.2: the service serves it, from a variable (R216) |
| 18 | No price words in ChatGPT's texts | New | GPT-07 | T71.4 | Open |
| 19 | Codex is among the surfaces a ChatGPT plugin is tested on | New | GPT-37 | §10; T71.6 | Open |
| 20 | The country read from the sign-in's address has no control the family can use | New | GPT-23 | T71.2 | Closed in T71.2: a tick in the profile (R220) |

### 8.1 The pace by address

- **The pace.** 20 requests a minute for each address, with a burst of 6 (R121), on both discovery documents and every `/oauth/` endpoint, `/oauth/token` among them. The address is the last of `X-Forwarded-For` (`internal/transport/http/middleware/address.go:66–74`).
- **Who calls.** The hosts' servers: Claude from `160.79.104.0/21`, 2,048 addresses; ChatGPT from 278 prefixes, some 36,000 addresses with one /17 among them, in the list of 22 September 2026.
- **What a family costs.** Claude renews "proactively up to five minutes before the stored expiry", and an access token lives 15 minutes, so a family in a lesson renews about every ten minutes: 0.1 a minute. A sign-in costs the host three requests, the two documents and the code's exchange, and four with a registration.
- **What one address carries.** 200 families in a lesson at the same moment, or six sign-ins a minute. Remark 60 counted 300 at one renewal in fifteen minutes.
- **What is not known.** How a host spreads its calls over its addresses. All from one address, 200 families at once is the ceiling; spread over N, it is 200·N. The directory's audience is not known either.
- **The options for T71.2.** Hold the published ranges to a pace of their own and keep each account's pace of renewals at Google (12 a minute, R127), which already bounds a single family; or keep 20 and count the refusals once listed; or raise the pace for every address.

### 8.2 What the health badge counts

| What MathTrail answers | How | The badge ("request errors") | The error rate of 30 days |
|---|---|---|---|
| A task refused by the checks, a stale request, arguments refused, the day's limit (`submittask.go:327–355`, `nexttask.go:243–259`) | An ordinary result with a `status` | No | No |
| A pace on a tool call, a failure of Drive or of the seal, a deadline, a busy solver, a file put back (`limits.go:81–98`, `failure.go:51–202`) | A result with `isError: true` | No | Yes |
| A call whose Google access is gone (R118) | `401` with a challenge, for every host but ChatGPT | Yes: "tool calls rejected for authentication problems" | No, though it counts in the total |
| Cloud Run's own errors and timeouts | HTTP `5xx` | Yes | Yes |

So the badge is at risk from lost Google sign-ins and the platform's own errors, not from the checks.

## 9. What the snapshot got wrong, and the advice that did not hold

**The snapshot of 2026-09-30** (RUN.md, T71.1):

| The snapshot | The pages on 2026-10-05 | Rows |
|---|---|---|
| Claude's review criteria hold the screenshots, `ui.domain`, the outbound range and the health indicator | The page is now the "Connector pre-submission checklist"; those moved to the pages of submission, MCP Apps, authentication and the listing | CL-11, CL-38, CL-36, CL-47 |
| After publication Claude's tools change without a new review | Still so; the tool names shown on the listing change by a listing edit a reviewer approves | CL-45 |
| ChatGPT's apps became plugins on 9 July 2026 | No official page read gives the date; only third-party sites do | — |
| A verified person, by an identity document and a selfie | "individual or business verification"; the help article (a search summary) asks for a government ID and, "if requested, a selfie" | GPT-01 |
| A justification of every tool's hint | OpenAI's pages disagree | GPT-12 |
| No timestamps or internal identifiers in results | "unless they are strictly required" | GPT-33 |
| At most two actions and no logo in the card | ChatGPT's guidelines are for "optional plugin UI"; it is Claude that makes its own binding | §4 |
| A review "from two weeks to several months" | "Review timelines may vary" | GPT-15 |
| A server's address never changes | ChatGPT's review requirements: the origin cannot, the path can, while its submission page sends a change of the address to support; Claude: not described | GPT-53, CL-46 |
| Anthropic suspends an account with signs of a child | Confirmed: "we'll disable accounts based on indicators of minor activity" | §2.1 |

**Gemini's advice**, where the task began:

| The advice | Verdict | Rows |
|---|---|---|
| Submit through a partner program | No: "there's no partner program to apply to first" at Anthropic; OpenAI takes a ZIP in a portal | CL-01, GPT-03 |
| A guest's way in without OAuth | No: the profile lives in the parent's Drive, which needs the parent's sign-in, and both platforms' reviewers are given an account | CL-09, GPT-09 |
| Keywords in the tools' descriptions for the directories' search, down to the OGE and the EGE | No: both want narrow, accurate descriptions, and OpenAI forbids "overly broad triggering"; MathTrail is for grades 1–6 | CL-16, GPT-28 |
| "Zero-Data Retention" in the reviewer's description | No: the service logs counts and a child's name of the month, and keeps the profile in the parent's Drive; a submission says what the privacy policy says | CL-20, GPT-21 |
| SSE | No: Claude asks for Streamable HTTP and will deprecate SSE; OpenAI's `mcp.json` names `"streamable-http"` | CL-31, GPT-03 |

## 10. Knowable only by submitting, publishing or a live run

| What | When it is learnt | Sooner by a live run, and its cost |
|---|---|---|
| Claude's seven acknowledgments, word for word | T71.5, in the portal | No |
| Whether a reviewer reads the package as "behavioral instructions from external sources" | The letter to Anthropic, or T71.5 | No |
| Whether a Community listing is held to the inline card's numbers | T71.5 | No |
| Whether the warning on a card's `ui/message` goes for a connector of the directory | After publication, T71.5 | No |
| Whether Claude asks the person before a tool marked destructive that a card calls, and before a card's write at all (remark 65) | After the release of T71.2, which marked `edit_profile` destructive by the documentation, with no live check (R215): the form and the choice of a topic on a card are to be looked at in Claude first thing | Yes: a build with a tool marked destructive, made in a scratch copy and never deployed, reached through a tunnel as a custom connector, as in T03 — some 10–20 messages of Claude's limit |
| Whether Claude keeps the widget's page by its address (remark 38) | T71.5 | In part, in the same run |
| What Claude does when a listed server's address changes | A letter to `mcp-review@anthropic.com` | No |
| The badge's numbers: lost sign-ins and the platform's errors | Thirty days after publication | No |
| Whether `openai/widgetDomain` meets ChatGPT's requirement while `ui.domain` is unset | T71.6, at upload and scan | No |
| Whether ChatGPT's domain check takes the token as the service serves it, in `text/plain` (R216) | T71.6, at upload | No |
| Whether a package stays under Claude Code's 25,000 tokens | Once Claude Code can sign in (CL-34) | Yes: count a package's tokens locally |
| Whether ChatGPT reads `_meta.securitySchemes` (remark 20) | T63 | T63 |
| What ChatGPT does with `ui/message` and `ui/open-link`, the task's parts as strings, the drawings | T63 | T63 |
| ChatGPT's widgets on phones | T71.6 (О-29) | No |
| Whether the plugin is held to Codex, and how its cases are judged there | T71.6 | No |
| ChatGPT's attestations, word for word | T71.6 | No |
| Whether OpenAI's identity verification needs a paid API account ([asked on OpenAI's forum](https://community.openai.com/t/can-a-remote-mcp-plugin-publisher-verify-identity-without-buying-api-credits/1403321) on 2026-10-05, unanswered) | T71.6 | No |
| Whether reviewers meet Google's check, if О-62 is A | T71.5, T71.6 | In part: a sign-in from another network and a new browser (T71.3) |
| The review times, the countries that list the plugin, enhanced distribution | T71.5, T71.6 | No |

## 11. Contradictions with PRODUCT, SPEC and the decisions

| Where | It says | The platforms say | Resolved by |
|---|---|---|---|
| PRODUCT 3 (`PRODUCT-V1.md:67`) | The child solves tasks "next to them or with their permission" | No account may be made available to anyone else | О-61 |
| PRODUCT 9.3 | Claude's submission "goes through a Team or Enterprise organization (per secondary sources)"; ChatGPT's rules at `apps-sdk/app-submission-guidelines` | Any paid plan (CL-01); the rules are at `plugins/plugin-guidelines` | This task rewrites 9.3 |
| PRODUCT 10 | "Directories: a privacy policy, minimal data collection, no advertising, a test account and examples for review" | Also identity verification, a support page, test cases, a recording and design rules | §7 |
| SPEC 7.1 | Every tool has `destructiveHint: false`, "nothing here reaches beyond the parent's own file" | CL-27, GPT-30 | T71.2: R215 |
| SPEC 8.1 | `_meta.ui.domain` "Not set" | ChatGPT requires it | T71.2: `openai/widgetDomain` beside it (R216) |
| SPEC 9.3 and remark 49 | A loopback redirect matches with its port | Claude Code needs any port | T71.2: R218 |
| SPEC remark 20 | `securitySchemes` in `_meta` alone | On the tool, `_meta` a mirror | T71.2 left it to T63 (R216) |
| SPEC remark 38 | Versioning the widget's address is the author's to decide | A cache key, kept for an hour | О-66 |
| R119 | "`get_profile` is annotated read-only and may still put a damaged file back" | Read-only means nothing changes; side effects are never hidden | T71.2: R214 |
| R191, R203 | The mark in the card's header | No logo in the response, in ChatGPT | О-63 |
| R193, R194 | After the trial series the row under a task ends with the topic's button | At most two actions at the bottom of the card | О-63 |
| R145 and the instructions | "Another task" sends its words "as the child's message" | §2 | О-61 |
| R208 | "a child who does not know says so in the chat" | §2 | О-61 |
| R198, R202 | The Coach page first in the menu | §5 | О-64 |
| R82 | On a deployment nobody signs in but through Google | A reviewer needs an account with no extra step | О-62 |
| `docs/live/03-first-deploy.md:24` and `docs/live/05-login-drive.md:133` | Google's app In production on 2026-09-22, Testing on 2026-09-30 | — | T71.3 |

## 12. What this changes in the plan

**RUN.md's reason for the order** (its item on the directories): it holds. After publication both platforms take new and changed tools without a new submission, with these qualifications. In Claude, a tool's name shown on the listing changes by a listing edit a reviewer approves (CL-45). In ChatGPT, a changed tool keeps its old definition until the new one passes the automatic checks, so the server has to accept the old schema meanwhile; a new tool waits for the checks, a deleted one goes at once, and the server's instructions and the widget's metadata are checked with the tools (GPT-50, GPT-51). A change of a tool after the submission, such as T80's of `submit_task`, therefore ships compatible with the definition before it. The address stays: ChatGPT's review requirements change an origin only by a new plugin, its submission page sends a change of the address to support, and Claude does not describe one (GPT-53, CL-46).

**T71.2**, the service (the platform in brackets):

- the annotations of every tool and R119's repair on a read, after the live check of §10, with an R entry whose text becomes ChatGPT's justifications (CL-27, GPT-30, GPT-32) [both];
- each description says what the tool does and its side effects, and rules of behaviour move where they fit (CL-16, GPT-28) [both];
- `answered_at`, `recent[].answered_at`, `age_seconds` with the seconds of an open request in `next_task`'s text, and `instructions_version` leave what the model sees (GPT-33) [ChatGPT, harmless to Claude];
- the widget's sandbox by one of the two ways of §7.3 (GPT-45); `openai/widgetCSP.redirect_domains` with the site (GPT-47); `securitySchemes` on the tool (GPT-48) [ChatGPT];
- the domain's token: from a variable on the service, or a file on the site at `mathtrail.app` (GPT-08); on the site, T71.6's step that puts it into a variable through Terraform changes with it [ChatGPT];
- the pace by address against the hosts' ranges (§8.1) [both];
- Claude Code's loopback redirect (CL-34) [Claude];
- the country read from the sign-in's address: a way for the family to turn it off, or the reason none is needed, recorded in R (GPT-23, R185) [ChatGPT];
- `serverInfo`'s `websiteUrl`, `description` and `icons`, optional (gap 1) [both];
- dropped from its list: `iss`, already sent, and `UserInfo`, not needed (gap 4).

Done in T71.2 on 5 October 2026, by R214–R220, with no live check: the hints by the documentation; the sandbox by `openai/widgetDomain` with `ui.domain` unset; the token served by the service; the pace kept at twenty; the country of the sign-in left out by a tick in the profile. Two items went elsewhere: the rules of behaviour in the descriptions to the task О-61 brings (gap 15), and `securitySchemes` on the tool itself to T63.

**T71.3**, the reviewer's access: the sign-in by О-62; the rest as it is written.

**T71.4**, the texts (the platform in brackets):

- the listing's texts within both platforms' limits (CL-04, GPT-04) [both];
- no price as a selling point in ChatGPT's texts (GPT-07) [ChatGPT];
- a support page over HTTPS and the documentation's address (CL-13, CL-21, CL-22, GPT-05, GPT-26) [both];
- Claude's form: use cases, company, data handling, the allowed link URI `https://mathtrail.app` (CL-05, CL-06, CL-08, CL-12) [Claude];
- five positive and three negative cases, release notes, a justification of every hint (GPT-10, GPT-12) [ChatGPT];
- three working examples and the screenshots' prompts (CL-24, CL-11) [Claude];
- the privacy policy against both platforms, in the version published when T71.4 runs (CL-20, GPT-21, GPT-23) [both];
- the icons, and where the ZIP lives and how it is built (GPT-03, GPT-06) [ChatGPT].

**New tasks, only on the author's answer:** the card's changes of О-63, a widget task before T71.4; the rewrite of О-61's option B, the instructions, the descriptions and the card's words, before T71.4; the site's menu of О-64, a change of the site before T71.4.

**T71.5 and T71.6**, proposals left to the author: T71.5 chooses the slug before submitting and crops each picture to the card, without the prompt; T71.6 treats the screenshots as optional, gives the recording's address, uses a project without EU data residency, verifies the identity of О-65 by what OpenAI asks (an identity document, and a selfie only if asked), puts the domain's token where T71.2 put it, chooses the countries, and finds out whether verification needs a paid API account.

## Appendix A. Draft letters

Written for option B of О-61 and for the card О-63 recommends, and **not sent**: the author sends them from `altedtech.info@gmail.com` once both are answered. Under option A of О-61 the second paragraph of each would describe the child answering in the chat beside the adult; if the card keeps its "Topic" button, question 3 of A.1 names it as a third button.

### A.1 To Anthropic

To `mcp-review@anthropic.com`, the address Anthropic gives for escalations about connectors ([Submission status](https://claude.com/docs/directory/submission-status), read 2026-10-05).

**Subject:** Before we submit: a maths connector used by a parent with a child

> Hello,
>
> We are preparing to submit MathTrail to the Connectors Directory: a free, open-source MCP connector with an MCP App, at `https://mcp.mathtrail.app/mcp` (site: https://mathtrail.app, code: https://github.com/MathTrail/mathtrail-standalone). Before we submit, we would like to ask three questions, so that we can change what needs changing rather than take your reviewers' time.
>
> MathTrail gives olympiad-style maths practice for grades 1 to 6. The account belongs to an adult, a parent or a tutor, who connects MathTrail, signs in with Google and keeps the child's profile, under a pseudonym, in their own Google Drive. In a lesson, the model speaks to the adult and the adult types every message; the child sits beside the adult and answers each task by pressing one of five options on the app's card. The child never types in the chat.
>
> 1. Is a connector used this way, with the adult present and typing and the child pressing answers on the card in the adult's chat, acceptable in the directory under the Consumer Terms and the Usage Policy?
> 2. To write each task, the model calls one of our tools, which returns a guide, reference tasks and formats: static text embedded in our server and public in our repository. With them come the child's grade, interests and up to 500 characters of the parent's notes, which the tool marks as information about the child, never as instructions. Does that count as directing Claude "to dynamically pull behavioral instructions from external sources" under section 2.F of the Software Directory Policy?
> 3. Our task card shows five answer options and, below them, two buttons: a hint and the next task. Do the inline card's "At most 2 actions" in the MCP Apps design guidelines count the answer options as actions?
>
> A yes or a no to each would be enough. Thank you.
>
> [Name], MathTrail
> altedtech.info@gmail.com

| Part of the letter | The fact | Where it is shown |
|---|---|---|
| Free and open source | MIT, a public repository | `LICENSE`, README |
| The adult owns the account, signs in with Google, keeps the profile in their Drive under a pseudonym | Only an adult signs in; the profile is one file in the parent's Drive | Privacy policy, "Who signs in" and "What is stored, and where"; 02-auth |
| The model speaks to the adult; the child presses answers and never types | True once option B is done | §2.4 |
| A tool returns a guide, reference tasks and formats, embedded and public, with the child's grade, interests and the parent's notes marked as information | `get_package` and its package; the notes are at most 500 characters | `content/package.go`, SPEC 4.1 |
| Five options and two buttons below them | The task card, once the choice of a topic has moved into the progress (О-63) | PRODUCT 3, scenario 2; §4 |

### A.2 To OpenAI

Through OpenAI's support at `help.openai.com`, which the review requirements name for questions before submission ([Review requirements](https://developers.openai.com/plugins/deploy/app-review), read 2026-10-05). No plugin exists yet, so there is no plugin ID to give.

**Subject:** Before we submit: a maths plugin a parent uses with a child

> Hello,
>
> We plan to submit MathTrail to the ChatGPT directory: a free, open-source plugin with an MCP server and a UI, at `https://mcp.mathtrail.app/mcp` (site: https://mathtrail.app, code: https://github.com/MathTrail/mathtrail-standalone).
>
> Its users are adults, parents and tutors. Its tasks are olympiad-style maths for children in grades 1 to 6, solved in a lesson the adult runs: the model speaks to the adult and the adult types every message, while the child, sitting beside them, answers on the card by pressing one of five options.
>
> 1. Is a plugin whose users are parents and tutors, and whose tasks are for children in grades 1 to 6, one that "explicitly targets children under 13" under the plugin guidelines?
> 2. Is a child under 13 pressing the answers on a card in a parent's ChatGPT, with the parent present and typing, within the Terms of Use?
> 3. Our card also runs in Claude, which refuses a `_meta.ui.domain` that is not its own. If we leave `_meta.ui.domain` unset and set `_meta["openai/widgetDomain"]`, does the plugin meet the requirement for a dedicated domain?
>
> A yes or a no to each would be enough. Thank you.
>
> [Name], MathTrail
> altedtech.info@gmail.com

| Part of the letter | The fact | Where it is shown |
|---|---|---|
| A plugin with an MCP server and a UI | The service and its widget | SPEC 7, 8 |
| Users are parents and tutors | Only an adult signs in | Privacy policy, "Who signs in"; PRODUCT 3 |
| The adult runs the lesson; the child answers on the card | True once option B is done | §2.4 |
| Claude refuses a foreign `ui.domain` | Claude's troubleshooting page | CL-38 |

## Appendix B. Sources

Every page was read on 5 October 2026: the Claude and OpenAI developer pages as their Markdown versions, through `https://claude.com/docs/llms.txt` and `https://developers.openai.com/plugins/llms.txt`; the help and legal pages as HTML. "Search summary" marks a page that refused fetching.

**Claude**

| Page | Address | Updated | How |
|---|---|---|---|
| Software Directory Policy | https://support.claude.com/en/articles/13145358-anthropic-software-directory-policy | 15 April 2026 | Page |
| Software Directory Terms | https://support.claude.com/en/articles/13145338-anthropic-software-directory-terms | 16 March 2026 | Page |
| Publish to the directory | https://claude.com/docs/directory/publish | — | Markdown |
| Submit a connector to the directory | https://claude.com/docs/connectors/building/submission | — | Markdown |
| Connector pre-submission checklist | https://claude.com/docs/connectors/building/review-criteria | — | Markdown |
| Track your directory submission | https://claude.com/docs/directory/submission-status | — | Markdown |
| Manage your directory listing | https://claude.com/docs/connectors/building/managing-your-listing | — | Markdown |
| Manage your listing after publishing | https://claude.com/docs/connectors/building/after-publishing | — | Markdown |
| Directory connectors vs custom connectors | https://claude.com/docs/connectors/building/directory-vs-custom | — | Markdown |
| Connector verification | https://claude.com/docs/connectors/verification | — | Markdown |
| Build an MCP server for Claude | https://claude.com/docs/connectors/building/index | — | Markdown |
| Authentication for connectors | https://claude.com/docs/connectors/building/authentication | — | Markdown |
| Get started with MCP Apps | https://claude.com/docs/connectors/building/mcp-apps/getting-started | — | Markdown |
| Design guidelines | https://claude.com/docs/connectors/building/mcp-apps/design-guidelines | — | Markdown |
| Open external links from MCP Apps | https://claude.com/docs/connectors/building/mcp-apps/external-links | — | Markdown |
| Troubleshoot MCP Apps | https://claude.com/docs/connectors/building/mcp-apps/troubleshooting | — | Markdown |
| Consumer Terms | https://www.anthropic.com/legal/consumer-terms | Effective 8 October 2025 | Page |
| Usage Policy | https://www.anthropic.com/legal/aup | Effective 15 September 2025 | Page |
| Age assurance on Claude | https://support.claude.com/en/articles/15171100-age-assurance-on-claude | 18 May 2026 | Page |
| Minimum age requirement access restriction | https://support.claude.com/en/articles/13117299-minimum-age-requirement-access-restriction | 16 March 2026 | Page |
| Guidelines for organizations serving minors | https://support.claude.com/en/articles/9307344-responsible-use-of-anthropic-s-models-guidelines-for-organizations-serving-minors | 16 March 2026 | Page |
| Child safety guidance for developers | https://support.claude.com/en/articles/15591275-child-safety-guidance-for-developers | 26 June 2026 | Page |
| Claude's client ID metadata document | https://claude.ai/oauth/mcp-oauth-client-metadata | — | The document |

**ChatGPT**

| Page | Address | Updated | How |
|---|---|---|---|
| Plugin guidelines | https://developers.openai.com/plugins/plugin-guidelines | — | Markdown |
| Upload and submit your plugin | https://developers.openai.com/plugins/deploy/submission | — | Markdown |
| Remote MCP server review requirements | https://developers.openai.com/plugins/deploy/app-review | — | Markdown |
| Plugin submission errors | https://developers.openai.com/plugins/deploy/submission-errors | — | Markdown |
| Reference | https://developers.openai.com/plugins/reference | — | Markdown |
| UI guidelines | https://developers.openai.com/plugins/concepts/ui-guidelines | — | Markdown |
| Authentication | https://developers.openai.com/plugins/build/auth | — | Markdown |
| Add UI to your MCP server | https://developers.openai.com/plugins/build/chatgpt-ui | — | Markdown |
| Build an MCP server | https://developers.openai.com/plugins/build/mcp-server | — | Markdown |
| Security & Privacy | https://developers.openai.com/plugins/guides/security-privacy | — | Markdown |
| IP egress ranges | https://developers.openai.com/api/docs/guides/ip-addresses | — | Markdown |
| ChatGPT connectors' addresses | https://openai.com/chatgpt-connectors.json | Created 22 September 2026 | The file |
| ChatGPT's client ID metadata document | https://chatgpt.com/oauth/client.json | — | The document |
| Terms of Use | https://openai.com/policies/row-terms-of-use/ | — | Search summary |
| API organization verification | https://help.openai.com/en/articles/10910291-api-organization-verification | — | Search summary |
| A question on verification without API credits | https://community.openai.com/t/can-a-remote-mcp-plugin-publisher-verify-identity-without-buying-api-credits/1403321 | Asked 5 October 2026 | Page |

**Google**

| Page | Address | How |
|---|---|---|
| Sign in with Google prompts | https://support.google.com/accounts/answer/7026266?hl=en | Page |
| Using OAuth 2.0 to access Google APIs (refresh tokens) | https://developers.google.com/identity/protocols/oauth2 | Page |
| Secret Manager pricing | https://cloud.google.com/secret-manager/pricing | Page |
