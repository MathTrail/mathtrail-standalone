# A letter on FERPA

As of 5 October 2026. A letter for a school that recommends MathTrail to families for use at home, about the Family Educational Rights and Privacy Act (FERPA). It is not signed. The notes say which fact each part of it rests on, and the words in brackets are for the maintainer to fill in. It describes the one model of use MathTrail offers, the one the [answers to the National Data Privacy Agreement](sdpc-ndpa-v2.md) are written for.

## Before signing

### The rules the letter relies on

From 34 CFR Part 99, as in force on 1 October 2026 (eCFR):

- **Education records** (§ 99.3) are records directly related to a student and maintained by an educational agency or institution, or by a party acting for the agency or institution.
- **Disclosure** (§ 99.3) means permitting access to, or the release, transfer or other communication of, personally identifiable information contained in education records.
- **A school official** (§ 99.31(a)(1)(i)(B)) may be an outside party to whom a school has outsourced institutional services or functions, provided the party performs a service the school would otherwise use employees for, is under the school's direct control with respect to the use and maintenance of education records, and is subject to § 99.33(a).
- **Redisclosure** (§ 99.33(a)) of personally identifiable information a school disclosed is allowed only with the prior consent of the parent or eligible student, and the receiving party may use it only for the purposes of the disclosure.

### What each part of the letter rests on

| Part of the letter | The fact | Where it is shown |
|---|---|---|
| The school creates no accounts and sends no data | The free app has no school or class accounts and no way to take in a roster | [The model of use](sdpc-ndpa-v2.md#the-model-of-use-families-at-home) |
| MathTrail sends the school nothing | No result, report or dashboard goes to anyone but the family | The same |
| A parent connects MathTrail and signs in; the child has no account | Only an adult signs in, with a Google account; the child answers on a card in the adult's chat | [Privacy policy](https://mathtrail.app/en/privacy/), "Who signs in" |
| MathTrail does not act for the school | It performs no service for the school, under no contract with it, and no school can direct how it keeps data | The model of use |
| A pseudonym, never a real name | The profile holds a pseudonym the adult chooses, and nothing asks for a name, birth date, school, address or photograph | [Data inventory](data-inventory.md#what-each-holds) |
| One file in the parent's Drive; no database | The service keeps no database and no disk, and reaches only the files it made | Data inventory |
| The logs: each answer under the account's code; 30 and 62 days; counts with no identifier; no public group under ten; Google's request log, with the sign-in's email | The fields of the lines, the retention of the two buckets, the rules of the public views, and what Cloud Run records of every request | Data inventory |
| No sale, no advertising, no training | Nothing is sold, shown as an advertisement or used to train a model | Privacy policy, "Advertising, profiling and training" |
| Download, delete, end access | The file is the parent's, and access is ended in the Google account | Privacy policy, "Taking your data back, and deleting it" |
| What reaches the chat's provider | Every tool's result is read by the chat's model | [Data inventory](data-inventory.md#what-the-chats-model-sees-tool-by-tool) |

### What has to be settled first

- Who signs, for which legal entity, and where letters are answered.
- That counsel has read the letter: its account of FERPA is MathTrail's understanding of the facts, not legal advice.
- That every fact in the table still holds on the day it is signed.
- That this package is on the repository's `main`, so that the link to the data inventory in the letter opens.

## The letter

> [Date]
>
> [Name], [title]\
> [School or district]
>
> **Re: MathTrail and the Family Educational Rights and Privacy Act**
>
> Dear [Name],
>
> You asked how MathTrail stands under FERPA if your school recommends it to families. This letter describes how MathTrail works when families use it at home, which is the only way its free app can be used.
>
> **How the school is involved.** Your school recommends MathTrail to families. It creates no accounts, sends us no roster and gives us no information about its students, and we send the school nothing back: no results and no reports. A parent who chooses to use MathTrail connects it to the parent's own Claude or ChatGPT and signs in with the parent's own Google account. The child has no account.
>
> **How we understand FERPA to apply.** FERPA's regulations define education records as records directly related to a student and maintained by an educational agency or institution, or by a party acting for it (34 CFR § 99.3). In the use described here, the school discloses no personally identifiable information from education records to MathTrail, and MathTrail does not act for the school. It performs no institutional service or function for the school and is not under the school's direct control, so we do not consider it a school official under 34 CFR § 99.31(a)(1)(i)(B). What a family makes with MathTrail is made by the family, for its own lessons, and stays in the family's control. This is our understanding of the facts, not legal advice, and your counsel may wish to confirm it.
>
> **How we handle a family's data.**
>
> - The child is known by a pseudonym the parent chooses. We never ask for a real name, a birth date, a school, an address or a photograph.
> - The child's profile is one file in the parent's own Google Drive. Our service has no database, and it reaches only the files it created, and only while the parent's permission lasts.
> - Our service's logs record what each request did and, for each answer, whether it was right, the mistake behind a wrong one, whether the hint was used and the child's chance of success, under a code for the parent's account. They never record the text of a task or the pseudonym. Their lines are kept for 30 days, and a copy of the lines children are counted from for 62 days. What we keep longer are counts that hold no identifier, and the counts we publish show no group of fewer than ten children. Beside them, the standard request log Google Cloud keeps for every service records, for 30 days, the address each request came from and the address it asked for, which during the sign-in can include the email address the parent's chat suggests.
> - We do not sell data, show advertising or use a child's data to train any model.
> - The parent can download or delete the profile at any time, and can end our access in the Google account's settings.
> - The lessons run in the family's own chat, so what our tools return there reaches the chat's provider under the family's own agreement with it.
>
> **If your school's use changed.** If your school wished to assign MathTrail as schoolwork, to collect results or to create accounts for students, what this letter says would no longer hold, and MathTrail's free app does not support that use.
>
> Our privacy policy is at https://mathtrail.app/en/privacy/, and a fuller account of what we keep, where and for how long is at https://github.com/MathTrail/mathtrail-standalone/blob/main/docs/grants/data-inventory.md. Questions about this letter can be sent to altedtech.info@gmail.com.
>
> Sincerely,
>
> [Name]\
> [Title], [legal entity]
