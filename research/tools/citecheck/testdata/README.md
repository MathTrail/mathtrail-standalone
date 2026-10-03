# Recorded registry answers

What Crossref (`crossref/`) and DataCite (`datacite/`) answered on 2026-09-25 for the DOIs the tests use, cut down to the fields `citecheck` reads. A file is named after its DOI in lower case, with each `/` written as `_`.

| File | The work, and why the tests need it |
|---|---|
| `crossref/10.1016_j.compedu.2011.02.003.json` | Klinkenberg et al. 2011, a journal article: an HTML entity in the journal's name |
| `crossref/10.1007_978-3-031-98414-3_23.json` | Gupta et al. at AIED 2025, a book chapter: two containers (series and proceedings), an accented name |
| `crossref/10.1177_1758835919874651.json` | A retracted article: its retraction reported twice, by Retraction Watch and by the publisher, and HTML tags in its title |
| `crossref/10.1088_1361-6595_aaebdb.json` | A corrected article: a correction does not put a work in doubt |
| `datacite/10.48550_arxiv.2503.16460.json` | The arXiv preprint of Gupta et al., which Crossref does not hold |
| `datacite/10.5281_zenodo.3554625.json` | Liu and Koedinger 2017, deposited on Zenodo: no container recorded |
