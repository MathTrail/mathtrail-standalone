def solve(options):
    parts = [2, 3, 4, 5, 6]  # each friend in turn eats 1/part of what is left
    shares = set()
    for slices in range(1, 1001):
        left = slices
        whole = True
        for part in parts:
            if left % part != 0:
                whole = False
                break
            left -= left // part
        if whole:
            common = gcd(left, slices)
            shares.add("%d/%d" % (left // common, slices // common))
    if len(shares) != 1:
        fail("the pizzas give %d different fractions" % len(shares))
    return match(options, list(shares)[0])
