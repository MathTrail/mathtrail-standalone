def solve(options):
    # Three favourite colours, with no shortage of children for any of them.
    limits = [20, 20, 20]
    # No colour is liked best by three children yet.
    failed = set()
    for counts in product(*[range(limit + 1) for limit in limits]):
        if max(counts) < 3:
            failed.add(sum(counts))
    for size in range(sum(limits) + 1):
        if size not in failed:
            return match(options, size)
    fail("no number of draws is ever enough")
