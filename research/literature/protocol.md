# Literature: how the research searches, reads and cites

One procedure for every literature task of the plan — S09, S10 and S16 (the literature behind the introduction and the ethics section), S31–S34 (the related work of paper A) — so that a search can be repeated by someone else and every citation can be checked.

## 1. What each area asks

| Task | Notes file | The questions it answers for paper A |
|---|---|---|
| S09 | `problem-solving.md` | What the research says about teaching problem solving through non-standard problems: Pólya, Schoenfeld, metacognition, productive struggle, desirable difficulties, transfer. Which of it grounds the corridor, the novelty of every task and the hint as a leading question. |
| S10 | `olympiad-and-measurement.md` | What olympiad-style practice is known to do, and how the literature measures problem-solving skill and the novelty of a task. |
| S16 | `privacy-law.md` | What may be collected about children, and what a study with children needs: COPPA and its 2025 amendments, GDPR article 8, the UK Age Appropriate Design Code, research ethics with minors. |
| S31 | `generation-verification.md` | How language models generate and check mathematics problems; how generated problems are judged; homogeneity and novelty of generated text. |
| S32 | `learner-models.md` | Rasch, 3PL, Elo and Glicko in adaptive practice; Urnings; the 85 % rule; mastery criteria; stochastic approximation. |
| S33 | `llm-tutors.md` | Language models as tutors, their measured benefits and harms; the study modes of the assistants; the products closest to MathTrail. |
| S34 | `mcp-safety.md` | The Model Context Protocol and MCP Apps; the safety of tools a model can call; the hosts' rules for children and their age limits. |

## 2. Where to search

Checked on 2026-09-25 from this network.

| Source | What it gives | Access from here |
|---|---|---|
| OpenAlex (`api.openalex.org/works?search=…`) | Search across disciplines; the works that cite a work (`filter=cites:<id>`) for forward snowballing; open metadata | answers |
| Crossref (`api.crossref.org/works?query.bibliographic=…`) | Search by bibliographic terms; the reference list a publisher deposited, for backward snowballing | answers |
| DBLP (`dblp.org/search/publ/api?q=…&format=json`) | Computer-science venues: AIED, EDM, L@S, ACL, NeurIPS | answers (it refused automated readers during S06) |
| arXiv listings and full texts (`arxiv.org/list/…`, `arxiv.org/html/<id>`) | Recent preprints; full text in HTML | answer; the arXiv API itself (`export.arxiv.org/api`) refuses this network with HTTP 406 |
| Semantic Scholar API | Search and citations | refuses without a key (HTTP 429); used when it answers |
| ERIC, ACM DL, SpringerLink, IEEE Xplore, Google Scholar | Education research and publishers' pages, through web search | pages answer; some full texts sit behind paywalls |

## 3. Queries and the search log

- Each task writes its queries into `search-log.tsv` before it runs them. One row per query: date, task, source, query, results, kept, note.
- A query starts from the area's questions (section 1). The task's own terms come first; synonyms and the names of known systems follow.
- The log keeps the queries that found nothing as well: a gap in the literature is a finding.

**Starting queries.** Each task begins with these, adds its own as it reads, and logs them all.

| Task | Queries |
|---|---|
| S09 | `mathematical problem solving heuristics instruction`; `Schoenfeld problem solving metacognition control beliefs`; `metacognitive instruction mathematics IMPROVE`; `productive failure mathematics meta-analysis`; `desirable difficulties learning`; `far transfer mathematics problem solving`; `adaptive expertise`; `inert knowledge` |
| S10 | `mathematics olympiad participation effects`; `mathematical circles enrichment outcomes`; `PISA creative problem solving mathematics`; `assessing transfer to novel problems`; `problem novelty assessment mathematics` |
| S16 | `COPPA 2025 amendments consent artificial intelligence`; `GDPR article 8 child consent information society services`; `Age Appropriate Design Code`; `research ethics minors online experiments education` |
| S31 | `math word problem generation large language models evaluation`; `mathematics question generation verification solvability`; `program-aided language models`; `homogeneity of language model outputs novelty`; `unanswerable math word problems`; `distractor generation mathematics misconceptions` |
| S32 | `Elo rating adaptive learning student model`; `Rasch model guessing parameter adaptive practice`; `Urnings rating system`; `eighty five percent rule optimal learning`; `mastery criterion consecutive correct knowledge tracing`; `Glicko rating educational` |
| S33 | `large language model tutor randomized trial learning outcomes`; `LLM mathematics tutoring errors evaluation`; `AI tutor overreliance students`; the study modes of ChatGPT, Claude and Gemini (web pages) |
| S34 | `Model Context Protocol security`; `MCP Apps specification`; `tool use language model agents prompt injection`; the hosts' age requirements (web pages) |

## 4. What is kept

- **Kept:** work that answers one of the area's questions, published in a peer-reviewed venue, or a preprint that such work cites or that is the primary source for a system the paper names.
- **Years:** any year for foundations (Pólya, Rasch, Robbins and Monro); from 2022 for language models.
- **Languages:** English; Russian for the olympiad tradition (S10).
- **Left out:** secondary retellings of a result when the primary source can be read; press releases, except as the primary source of a product fact (a study mode, a host's rule), which is then cited as a web page (section 7).
- **Snowballing:** backward through a kept work's references (Crossref, or the paper itself); forward through the works that cite it (OpenAlex). Two rounds at most, logged like any query.

## 5. How a work is read

- **Full text first.** arXiv's HTML (`arxiv.org/html/<id>`), PubMed Central, or the publisher's page. A PDF is read by extracting its text locally for reading only; the text is never committed. Where the publisher refuses automated readers, a copy the authors or a repository posted will do: the notes name the copy, and say when its page numbers are not the published ones.
- **Summaries are not quotes** (Q54). The tool that fetches a page answers with a summary made by a small model. A passage is quoted only from the original, and the notes keep it verbatim with its page or section.
- **Paywalls.** A work read from its abstract and metadata only is marked so in the notes and is cited for nothing beyond its abstract.
- **Numbers** a paper takes from a work are checked against the original and go into a data file with the passage's location (rule G2).

## 6. The notes of an area

`literature/<area>.md`, in English, one section per work:

- the key in `refs.bib`;
- what the work did and found, in our words;
- the passage each claim rests on, verbatim, with its page or section, and how it was read (full text, abstract only);
- what paper A takes from it, and how MathTrail differs.

The file ends with the area's answers to its questions and a "Remarks" section for what stays open.

## 7. The bibliography

- **`refs.bib`** holds every work with a DOI or an arXiv id that the papers or notes cite. Entries are written only by `just research citecheck add <DOI | arXiv:id>`, from the registry's record; `just research citecheck` checks them all again. An author may change an entry's type, abbreviate each word of its venue ("Comput. Educ." for "Computers & Education"), or write an accented name with LaTeX commands (`M{\"u}ller`), or set a title's mathematics in math mode again, and the check must stay green. A run in which a registry does not answer ends with status 2 rather than 1: the entries it could not ask about are not checked, not wrong.
- **Keys** are the first author's family name, the year and the first title word that says something: `klinkenberg2011computer`. A second work with the same key gets a letter: `gupta2025beyondb`. A work whose record names no author, as with the reports of the National Research Council, is keyed `anonymous`: `anonymous2001adding`.
- **What the check compares:** title (subtitle included or not), the authors' family names in order ("and others" allowed after the first ones; an organisation is braced, as `add` writes it), the year (any of the years the registry records: of issue, in print, online) and the journal or proceedings, word by word. It fails an entry whose work a registry marks retracted, withdrawn, removed, partly retracted or under an expression of concern.
- **What it cannot see.** Crossref and DataCite are asked; arXiv's own API refuses this network, so preprints are checked by the DOI arXiv registers with DataCite (`10.48550/arXiv.<id>`). A registry that records no journal or proceedings — DataCite records none for preprints and deposits — cannot confirm one, so such an entry may not claim a venue: a preprint that was later published is cited by the published version's own DOI. Authors are compared only when the record names some: for a record without them, as for the reports of the National Research Council, an entry's authors go unchecked. arXiv's withdrawals are not in DataCite's record; the reading task sees them on the abstract page.
- **Web pages** without a DOI — a host's age rule, a study mode's announcement — go into `web.bib` with the date they were read, and the notes keep the passage. `citecheck` does not check them.
- **`seed.bib`** keeps the prototype's leads. A lead becomes a citation only when a literature task reads the work and adds it to `refs.bib`.
