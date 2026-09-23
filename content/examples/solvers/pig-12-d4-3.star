def solve(options):
    # The stock: red, blue.
    limits = [6, 6]
    # Some colour has fewer than two.
    failed = set()
    for counts in product(*[range(limit + 1) for limit in limits]):
        if min(counts) < 2:
            failed.add(sum(counts))
    for size in range(sum(limits) + 1):
        if size not in failed:
            return match(options, size)
    fail("no number of draws is ever enough")
