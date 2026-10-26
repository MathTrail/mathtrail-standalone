def solve(options):
    # Every way the five could be knights or liars: the islander at place k
    # says that at least k of the five are knights, which is true for a
    # knight and false for a liar.
    counts = set()
    for kinds in product([True, False], repeat=5):
        knights = len([k for k in kinds if k])
        if all([(knights >= place + 1) == kinds[place] for place in range(5)]):
            counts.add(knights)
    if not counts:
        return match(options, "None: the five sentences cannot all fit")
    if counts == set(range(6)):
        return match(options, "Any number from 0 to 5")
    if len(counts) == 1:
        return match(options, "Exactly %d" % list(counts)[0])
    fail("%d numbers of knights fit" % len(counts))
