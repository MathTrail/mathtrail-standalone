def solve(options):
    # Four colours, ten of each.
    limits = [10, 10, 10, 10]
    # No two of one colour yet.
    failed = set()
    for counts in product(*[range(limit + 1) for limit in limits]):
        if max(counts) < 2:
            failed.add(sum(counts))
    for size in range(sum(limits) + 1):
        if size not in failed:
            return match(options, size)
    fail("no number of draws is ever enough")
