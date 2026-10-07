# Answers to the SDPC National Data Privacy Agreement

As of 5 October 2026. The National Data Privacy Agreement (NDPA) is the Student Data Privacy Consortium's standard agreement between a school district in the United States and a provider of an online service; these answers follow its version 2.2. They are written for one model of use, the only one MathTrail offers: families using it at home on a school's recommendation. MathTrail has signed no agreement. The answers are here for a school that asks, and every row marked for the maintainer has to be settled before anything is signed.

The agreement's text belongs to the Access 4 Learning Community, which lets only its members use it and forbids altering it in substance. This page reproduces none of it and refers to its sections by number.

## The model of use: families at home

- **The school recommends MathTrail** to families, in a letter home for instance. It creates no accounts, hands over no roster and gives MathTrail no data about any student.
- **A parent who chooses to use it** connects MathTrail to the parent's own Claude or ChatGPT, signs in with the parent's own Google account and runs the lessons at home. The child has no account.
- **What the lessons produce stays with the family**, in the family's own Google Drive. MathTrail sends the school nothing: no results, no reports, no dashboard. The free app has no school or class accounts.
- **Whether an agreement is needed at all** under this model is the school's call. The agreement's Student Data is data that describes a student and is gathered for a school purpose; whether a family's own lessons at home are one is for the school and its counsel to judge.
- **What would change these answers** is the school assigning lessons as homework, collecting results or creating accounts. That is another model of use, and the free app does not support it.

## Exhibit A: products and services

- **MathTrail**, a free, open-source app of olympiad-style maths for grades 1–6. It is used inside the family's own Claude or ChatGPT as an MCP app and is served at `mcp.mathtrail.app`.
- **Its site**, [mathtrail.app](https://mathtrail.app), which sets no cookies, runs no analytics and loads nothing from other domains. GitHub, which serves it, keeps its own access log.

## Exhibit B: schedule of data

Under this model the school provides no Student Data. The table says what MathTrail itself handles for a family that uses it, in the agreement's categories: **R** is needed for the lessons, **O** is optional and the adult may leave it out, **Log** is never asked for but lands in Cloud Run's request log when a request carries it, and **No** is never collected. Where each item is kept and for how long is in the [data inventory](data-inventory.md).

| Category | Element | | What exactly |
|---|---|---|---|
| Application technology metadata | IP addresses of users, use of cookies | R | At every sign-in the adult's address is looked up for its country, which the tokens and the log lines then carry. The service writes the address nowhere: it is held only in the memory of the instance that answers, as the key of a counter of the pace of requests. Cloud Run's request log records it for 30 days. Two cookies, needed for the sign-in alone |
| | Other application technology metadata | R | An opaque code for the adult's account, and a random code for each task, which joins up the calls it took, in the service's log |
| Application use statistics | Metadata on user interaction with the application | R | What each request did, with, for each answer, whether it was right, the trap behind a wrong one, whether the hint was used and the child's chance of success; on three kinds of line, a name the child is counted under for one calendar month. In the log for 30 days, and a copy for 62 |
| Assessment | Other assessment data | R | The child's answers to the tasks: right or wrong, the trap behind a wrong one, whether the hint was used; in the profile, and one by one in the log |
| | Standardized test scores, observation data, voice recordings | No | |
| Attendance | All | No | |
| Communication | Online communication captured | No | MathTrail is not given the conversation, only what its tools are called with |
| Conduct | Conduct or behavioral data | No | |
| Demographics | Date of birth, place of birth, gender, ethnicity or race | No | |
| | Language information | O | The language of the lessons, if the adult chose one; it is not asked as the child's native language |
| | Other demographic information | O | The country and, in the United States, the state the family lives in, if the adult gives them. The country the adult signed in from is under the first row |
| Enrollment | Student grade level | R | The grade, 1 to 6 |
| | School enrollment, homeroom, guidance counselor, curriculum programs, year of graduation | No | |
| Parent/guardian contact information | Address, phone | No | |
| | Email | Log | Never asked for. A chat may put an email address in the sign-in's address, which Cloud Run's request log then records for 30 days |
| Parent/guardian ID | Parent ID number | R | The opaque code for the adult's account, made with the service's key from the Google account; in the log it ties the lines of one account together, its children's included |
| Parent/guardian name | First and/or last | No | |
| Schedule | All | No | |
| Special indicator | All | O | None is asked for. The adult's notes are optional free text of up to 500 characters and hold whatever the adult writes, which may include such information; they are kept in the profile and reach the chat's model |
| Student contact information | All | No | |
| Student identifiers | Local ID number, state ID number | No | |
| | Provider/app assigned student ID number | R | A random identifier of the child, in the profile; never in a log |
| | Student app username | R | The pseudonym the adult chooses, never the real name |
| | Student app passwords | No | The child has no account |
| Student name | First and/or last | No | |
| Student in app performance | Program/application performance | R | Ratings by topic, topics mastered, a summary of past answers and the traps that came up, in the profile; and each answer, as the row of application use statistics describes it, in the log |
| Student program membership | All | No | |
| Student survey responses | All | No | |
| Student work | Student generated content, other student work data | No | The child picks one of five options; nothing is written or drawn |
| Transcript | All | No | |
| Transportation | All | No | |
| Other | Other data collected | O | In the profile: the child's interests, the skills kept out of tasks, the adult's notes and the topic the lessons are kept to; and the service's own records there: a sketch of each of the last 200 tasks, the current task with its answer sealed, and the count of today's tasks |

**Where it is stored (section 5.1).** The copy of the lines children are counted from and the counts are kept in Google Cloud's `us-central1`, in the United States. The service's log is in a bucket whose location is `global`, and the keys are replicated automatically: Google chooses where both are kept, which may be outside the United States, and no list of countries can be given for them. The profile is wherever Google keeps the family's Drive, and the conversation wherever the chat provider keeps it.

## The standard clauses

| Section | Answer | Why |
|---|---|---|
| 1.1 School official | Not applicable | MathTrail performs no service for the school and receives nothing from it, and it has no school accounts through which a school could direct it |
| 1.2 Products and services | Holds | Listed in Exhibit A above; a new one would be added through the addendum the agreement provides |
| 1.3 Student data | Holds | Listed in Exhibit B above; a new element would be added through the same addendum |
| 2.1 Data property of the school | Holds in substance | MathTrail claims no right in a family's data: the profile is the family's file, in the family's Drive. Nothing the school owns reaches MathTrail |
| 2.2 Access by parents | Holds by design for the profile; the log is the exception | The parent opens, downloads, changes or deletes the profile in the parent's own Drive at any time, and MathTrail holds no other copy of it. A request that reaches MathTrail is pointed to that file rather than to the school. The service's log is the exception: for up to 62 days it keeps a line for each answer, under the code of the parent's account, which the parent can neither see nor delete, and which leaves on its own |
| 2.2.1–2.2.2 Export | Holds by design | The profile is in the family's control from the start |
| 2.3 Subprocessors | Partly; for the maintainer | MathTrail's processors are Google — Cloud Run, Cloud Logging, BigQuery, Secret Manager, Cloud Trace and Cloud Monitoring, and the sign-in and Drive on the family's own Google account — and GitHub, which serves the site. Their terms are click-through agreements, which the agreement accepts. Whether each is no less strict than the agreement and forbids selling the data is still to be checked against their current terms. The chat provider is the family's own choice and contract, not MathTrail's subprocessor |
| 3.1–3.4 Duties of the school | The school's | |
| 4.1 Compliance with laws | For the maintainer | A commitment to make with counsel. The practices it would rest on are in the [privacy policy](https://mathtrail.app/en/privacy/) and the [data inventory](data-inventory.md) |
| 4.2 Authorized use | Holds by design | The data is used to run the lessons and for nothing else. The log lines keep the service running and are counted into totals that hold no identifier |
| 4.3 Employees | Holds today | MathTrail is run by one person and has no employees. A confidentiality agreement for anyone given access later is for the maintainer to confirm |
| 4.4 No disclosure | Holds by design | Nothing is sold or handed to anyone. What the chat's model sees reaches the provider the family chose for its own chat. An order to disclose could reach only what the logs hold, for up to 62 days, and the counts: the service holds no profile |
| 4.5 De-identified data | Holds in practice; one check open | The counts hold no identifier. The public views show closed months only and no group of fewer than ten children; they round every count to the nearest five, give shares to three decimal places and means to one, and never name a school. Nothing is done to re-identify anyone. The counts serve to show how the service is used, which is among the purposes the section allows. The section asks that de-identification follow NIST's standards or the US Department of Education's guidance, and checking the method against that guidance is still to be done. Part of that check: shares and means are worked out from the exact counts, so in a small group they can give back a count the rounding hides |
| 4.6 Disposition | Holds by design, with two limits | A family deletes the file, and any file set aside beside it, and MathTrail keeps no copy of the profile. What the chat's model saw stays with the chat provider, under the provider's policy. Log lines, the line of each answer among them, cannot be deleted one by one: they leave 30 days after they were written, and the copy 62 days after, which can be up to two days past the 60 the section gives after a request. The [data inventory](data-inventory.md) is the schedule of retention |
| 4.7 Advertising | Holds by design | No advertising, no profile of a child or a family for any other purpose, nothing sold, no model trained on the data. MathTrail sends no product recommendations or news, and holds no email address to send them to. The parent, not the school, signs in and sets up the profile, so MathTrail never relies on a school's consent given on a parent's behalf |
| 5.1 Storage outside the US | Answered | Under Exhibit B above |
| 5.2 Security audits | Not yet; for the maintainer | No external audit has been made. What exists: the [OpenSSF Best Practices](../openssf-best-practices.md) badge at the passing level, and the checks every change passes — CodeQL, `gosec`, `govulncheck`, `npm audit`, a secret scan of the whole history and a scan of the runtime image. Whether to commit to a yearly assessment is for the maintainer |
| 5.3 Data security | Partly; for the maintainer | The safeguards are in the [data inventory](data-inventory.md) and the [security policy](../../SECURITY.md). None of the frameworks of Exhibit F has been adopted or assessed against, so none can be marked; whether to adopt one, and which, is for the maintainer. Security concerns go to a private report through GitHub, as the security policy describes |
| 5.4 Data breach | Partly; for the maintainer | What is done is in the [data inventory](data-inventory.md#when-a-key-leaks): the cause fixed and delivered, the keys replaced, a public notice on the site, in the README and as a GitHub security advisory. What the section asks beyond that — notice to the school within 72 hours with the items it lists, and a written plan for responding, summarized on request — is for the maintainer to commit to. Under this model MathTrail does not know which families a school recommended it to, or who they are, so it could not name the people affected to a school; the public notice is how families learn |
| Contract terms | For the maintainer | Term, notices, governing law and venue, change of control and successors are for the maintainer to settle with counsel |

## Exhibits D to G

- **D, instructions for disposition.** A school fills it in later, if ever, and the agreement forbids filling it in at signing. Under this model there is no Student Data of the school's to dispose of.
- **E, the general offer of privacy terms.** Signing it would offer the same terms to every other school that accepts them. For the maintainer to decide.
- **F, cybersecurity frameworks.** Under section 5.3 above.
- **G, state terms.** They depend on the school's state, and the Consortium publishes them; they are read when a school of that state asks.

## What a school would still need

- The provider's legal name and address, a designated representative and a contact for security notices.
- The maintainer's answers on the rows marked for the maintainer above: the terms of the subprocessors, compliance with laws, confidentiality, a yearly assessment, a framework, the commitments on a breach, the general offer and the contract terms.
