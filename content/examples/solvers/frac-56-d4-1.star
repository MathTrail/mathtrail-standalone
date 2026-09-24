def solve(options):
    # Every library in which each share is a whole number of books gives the
    # same fraction, written in lowest terms as the options write it.
    shares = set()
    for books in range(1, 1001):
        if books * 2 % 5 != 0:
            continue
        novels = books * 2 // 5
        others = books - novels
        if novels % 2 != 0 or others % 6 != 0:
            continue
        hard = novels // 2 + others // 6
        common = gcd(hard, books)
        shares.add("%d/%d" % (hard // common, books // common))
    if len(shares) != 1:
        fail("the libraries give %d different fractions" % len(shares))
    return match(options, list(shares)[0])
