def solve(options):
    # Lots of each colour: red, blue, green.
    limits = [20, 20, 20]
    # No two balls of one colour yet.
    failed = set()
    for counts in product(*[range(limit + 1) for limit in limits]):
        if max(counts) < 2:
            failed.add(sum(counts))
    for size in range(sum(limits) + 1):
        if size not in failed:
            return match(options, size)
    fail("no number of draws is ever enough")
