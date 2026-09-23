def solve(options):
    # The stock: even numbers, odd numbers.
    limits = [5, 5]
    # No even number yet.
    failed = set()
    for counts in product(*[range(limit + 1) for limit in limits]):
        if counts[0] < 1:
            failed.add(sum(counts))
    for size in range(sum(limits) + 1):
        if size not in failed:
            return match(options, size)
    fail("no number of draws is ever enough")
