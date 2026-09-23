def solve(options):
    # Four colours, "many" of each: ten is more than any draw here needs.
    limits = [10, 10, 10, 10]
    # Fewer than two complete pairs yet.
    failed = set()
    for counts in product(*[range(limit + 1) for limit in limits]):
        if sum([n // 2 for n in counts]) < 2:
            failed.add(sum(counts))
    for size in range(sum(limits) + 1):
        if size not in failed:
            return match(options, size)
    fail("no number of draws is ever enough")
