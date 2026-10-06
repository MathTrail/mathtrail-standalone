# MathTrail, paper A: the artifact

This is the artifact of the paper *The Model Writes, the Service Checks: Olympiad-Style Maths Problems for Primary-School Children inside Chat Assistants*. It holds the service at the commit the paper describes, with its reference tasks, and everything the paper's numbers come from: the experiments' code, protocol, data and results, and the evidence ledger.

## What is where

| Path | What it holds | In the paper |
|---|---|---|
| `service/` | The service's repository as it stands at the commit every claim of the paper is pinned to, the one `service/research/evidence/product-stats.txt` names, with the research below in place of the research plan of that day | Sections 3–5 |
| `service/content/examples/` | The reference tasks, each with its solver and a trap behind each wrong option | Section 3 |
| `service/content/catalogs/` | The topics, the traps and the skills | Section 3 |
| `service/research/experiments/PROTOCOL-A-offline.md` | The protocol of the experiments, frozen before their first run; the `.ots` file beside it is its OpenTimestamps proof, and `DEVIATIONS.md` lists where the runs and the paper depart from it | Section 6 |
| `service/research/experiments/faultinject/` | E-A1, defects injected into the reference tasks: the code, and in `results/` the data the paper cites | Section 6.1, Table 2 |
| `service/research/experiments/learnersim/` | E-A3, simulated learners under the service's update and the baselines | Sections 5 and 6.2, Figure 2 |
| `service/research/experiments/perf/` | E-A4, what a review costs and how large the model's package is | Section 6.3 |
| `service/research/experiments/reviewing/` | The review harness the experiments share | — |
| `service/research/evidence/ledger.md` | Every claim the paper makes about the service, with the file and line in this tree that prove it; `claims-product.json` is its source, and `product-stats.txt` holds the facts computed from the code by `product.sh` | throughout |
| `service/research/evidence/prototype.md` | The claims about the prototype the service grew from, each proven by a quotation from the prototype's public repository | Sections 3 and 6 |
| `service/research/paper-a/numbers.tex` | Every number the paper prints, as the macro it is printed through | throughout |
| `service/research/literature/numbers.txt` | The numbers the paper takes from other works, each with the passage it comes from | Sections 1, 2, 5 and 8 |

## Reproducing the results

You need Go at the version `go.mod` names. The first build downloads the Go modules `go.sum` lists. Every random choice comes from a fixed seed, so E-A1 and E-A3 reproduce byte for byte; the times E-A4 measures depend on the machine, the sizes it measures do not.

```sh
cd service/research
# No Go workspace outside the artifact may stand in for its modules, and the
# artifact is no git checkout to stamp a build with.
export GOWORK=off GOFLAGS=-buildvcs=false

# E-A1, Section 6.1 and Table 2: about two minutes
ea1=$(mktemp -d) && cp experiments/faultinject/results/reading.csv "$ea1/"
go run ./experiments/faultinject -out "$ea1"
diff -r -x provenance.txt experiments/faultinject/results "$ea1"

# E-A3, Sections 5 and 6.2 and Figure 2: about four minutes
ea3=$(mktemp -d)
go run ./experiments/learnersim -out "$ea3"
diff -r -x provenance.txt experiments/learnersim/results "$ea3"

# E-A4, Section 6.3: the times are this machine's, the package sizes are not
ea4=$(mktemp -d)
go run ./experiments/perf -out "$ea4"
diff experiments/perf/results/packages.csv "$ea4/packages.csv"
grep '^package' experiments/perf/results/numbers.txt > "$ea4/shipped-package-lines"
grep '^package' "$ea4/numbers.txt" | diff "$ea4/shipped-package-lines" -
```

`reading.csv` holds verdicts on a sample of the out-of-scope cases: whether each case is the defect its operator is named after. No program makes them: they were recorded by the AI assistant that ran the experiment, and the paper does not report them. The run reads them from its output directory and adds them to its tables; without the file it leaves those entries out, and everything else comes out the same.

`provenance.txt` beside each result names the commit the run was made on and the Go and Starlark versions it used.

Each results directory has a `numbers.txt` of `key=value` lines, and `paper-a/numbers.tex` defines one macro per line: `\stat{ea1}{key}` is E-A1's `key`, `ea3` E-A3's and `ea4` E-A4's; `product` is a fact of `evidence/product-stats.txt`, `ledger` one of `evidence/ledger-numbers.txt`, and `literature` one of `literature/numbers.txt`, all under `service/research/`.

## What is not here

- **Data about children.** None was collected: the experiments run on the reference tasks and on simulated learners.
- **Model outputs.** Neither experiment calls a language model; the one file in the results a model wrote is `reading.csv`, above.

## Licence

MIT: the service, its reference tasks and catalogs (`service/LICENSE`), and the research code and data (`service/research/LICENSE`). The Go modules the service links are listed in `service/THIRD_PARTY_LICENSES`, and those the experiments add in `service/research/THIRD_PARTY_LICENSES`.
