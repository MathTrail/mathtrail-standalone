def solve(options):
    # The ten possible last digits; two of each is more than any choice here needs.
    limits = [2, 2, 2, 2, 2, 2, 2, 2, 2, 2]
    # No two numbers ending in the same digit yet.
    failed = set()
    for counts in product(*[range(limit + 1) for limit in limits]):
        if max(counts) < 2:
            failed.add(sum(counts))
    for size in range(sum(limits) + 1):
        if size not in failed:
            return match(options, size)
    fail("no number of draws is ever enough")
