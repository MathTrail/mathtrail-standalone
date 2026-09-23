def solve(options):
    # The stock: red, blue, green.
    limits = [3, 3, 3]
    # Fewer than two colours drawn so far.
    failed = set()
    for counts in product(*[range(limit + 1) for limit in limits]):
        if len([n for n in counts if n > 0]) < 2:
            failed.add(sum(counts))
    for size in range(sum(limits) + 1):
        if size not in failed:
            return match(options, size)
    fail("no number of draws is ever enough")
