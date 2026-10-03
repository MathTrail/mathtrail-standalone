# Paper A: what submitting it takes

Task S66 of [RUN.md](../RUN.md). Paper A goes to AIED 2027, main track, full paper (Q58). AIED 2027 had published no call for papers by 2026-10-02, so the requirements below are the AIED 2026 call's, read that day ([call](https://aied-conference.org/2026/call-for-paper)); each one is read again in the 2027 call before the paper goes in. The expected dates are in [venues.md](../literature/venues.md): abstract about late January 2027, paper about early February.

## The venue's requirements

| Requirement, as the 2026 call states it | How paper A meets it | Status |
|---|---|---|
| "Full papers · 14 pages (including references)"; "All submissions must be in Springer format" | The anonymous build is 14 pages with its references, in Springer's `llncs` class (Q67) | met |
| Anonymisation: "Eliminate all information that could lead to their identification (names, contact information, affiliations, patents, names of approaches, frameworks, projects and/or systems)" | The text writes `\System` and `\ArtifactURL`; the build refuses the name or the repositories typed by hand, and a test document built both ways shows the name and the commit absent from the anonymous build (Q67) | met |
| "Cite own prior work (if needed) in the third person" | The paper cites no earlier work of its authors | met |
| "Eliminate acknowledgments and references to funding sources" | The credits block is printed in the named build only | met |
| Responsible reporting: "Describe clearly the composition of human-sourced data, including participant demographics", the imbalances, the ethical issues, inclusive language | No human-sourced data: no child and no participant took part. The ethics paragraph of Section 8 covers children's data, consent and the hosts' age rules | met |
| "Submissions must follow Springer policies on publication (including policies on the use of AI in the authoring process)" | Springer asks for the use of a language model to be documented in the methods ([venues.md](../literature/venues.md)). The paper does not disclose it, by the author's decision (Q83, Q86) | the author's decision, against the policy |
| Preprints "remain acceptable (provided that they follow Springer's policies)" | A preprint on arXiv needs a personal endorsement for a first-time author (K09) | the author's call |
| "Work must be original and not under consideration in other journals or conferences" | EDM and L@S close in the same weeks; one venue is chosen at submission ([venues.md](../literature/venues.md)) | at submission |
| "Abstract submission is mandatory"; "Submissions are handled via EasyChair" | Registering the abstract and the paper is the author's | at submission |

## What the paper itself still needs

| Item | Where | Status |
|---|---|---|
| No TBD left: `just research paper-a-final` refuses to finish while one is | The anonymous text has one, the anonymised artifact link, which `paper-a/preamble.tex` takes once it exists (K11); the named version keeps the artifact's DOI (K11) and its credits (K05) | waits for K11 and K05 |
| The abstract within the 150–250 words the LNCS template asks for ([template](https://www.overleaf.com/latex/templates/springer-lecture-notes-in-computer-science/kzwwpvhwnvfj)) | Abstract | met (S64) |
| An anonymous artifact link for the review | `\ArtifactURL` of the anonymous build; the artifact is `just research release anonymous` (Q84), hosted where K11 decides | waits for K11 |
| The credits of the final version: CRediT roles, acknowledgements and funding, competing interests | The named build's credits block (K05) | waits for K05 |

## Right before submitting

1. Read the 2027 call and update the table above.
2. `just research citecheck` and `just research ledger-check`.
3. Read again the pages the related work cites for products that change fast: Tutor MCP, MathForces and Beast Academy (S62).
4. `just research release anonymous` and `just research release-check anonymous`, and upload the tarball to the host K11 chose.
5. `just research paper-a-final`, and read every page of `paper-a/build/paper-a-submission.pdf`.
