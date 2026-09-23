def solve(options):
    # The stock: red left, red right, blue left, blue right.
    limits = [4, 4, 4, 4]
    # No matching pair of one colour yet.
    failed = set()
    for counts in product(*[range(limit + 1) for limit in limits]):
        red = counts[0] > 0 and counts[1] > 0
        blue = counts[2] > 0 and counts[3] > 0
        if not red and not blue:
            failed.add(sum(counts))
    for size in range(sum(limits) + 1):
        if size not in failed:
            return match(options, size)
    fail("no number of draws is ever enough")
