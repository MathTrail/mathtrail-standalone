# This copy is anonymised

This copy of the artifact is the one the paper's review version links to. It differs from the copy published with the final paper in three ways.

- **Names.** The system's name and the name of its prototype's repository are replaced throughout, in file contents and in Go module paths alike, so the code still builds and runs.
- **Commits.** Every commit hash of the system's repository and of its prototype's is masked with zeros.
- **What is left out.** Of the service, this copy keeps the packages the experiments and their tests import: {{PACKAGES}}. Of these, `service/content/` holds the reference tasks and the catalogs. Its server, deployment, site and documents are in the named copy, and so are the files some claims of the evidence ledger point to.

Two checks can be made only on the named copy:

- The OpenTimestamps proof `service/research/experiments/PROTOCOL-A-offline.md.ots` attests the protocol as it was written, whose SHA-256 is `{{PROTOCOL_SHA256}}`. The copy here differs from it only where names and commits are replaced.
- The claims about the prototype quote files of its public repository, whose name is replaced here.

Every experiment reproduces from this copy as `README.md` says.
