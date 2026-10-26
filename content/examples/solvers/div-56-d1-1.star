def solve(options):
    # Every number that gives 5 with 3 left over when divided by 7; four times
    # it is divided by four times 7.
    results = set()
    for n in range(0, 1001):
        if n // 7 == 5 and n % 7 == 3:
            results.add("%d with %d left over" % ((4 * n) // 28, (4 * n) % 28))
    if len(results) != 1:
        fail("%d results fit" % len(results))
    return match(options, list(results)[0])
