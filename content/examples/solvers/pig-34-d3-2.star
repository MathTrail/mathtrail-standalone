def solve(options):
    # The stock: red, blue, green.
    limits = [5, 5, 5]
    # Fewer than two red balls yet.
    failed = set()
    for counts in product(*[range(limit + 1) for limit in limits]):
        if counts[0] < 2:
            failed.add(sum(counts))
    for size in range(sum(limits) + 1):
        if size not in failed:
            return match(options, size)
    fail("no number of draws is ever enough")
