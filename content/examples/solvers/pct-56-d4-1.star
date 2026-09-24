def solve(options):
    first = []
    for apples in range(1, 1001):
        if apples * 25 % 100 != 0:
            continue
        red = apples * 25 // 100
        if red < 6:
            continue
        # After 6 red apples are eaten, 10% of the apples left are red.
        if (red - 6) * 100 == (apples - 6) * 10:
            first.append(apples)
    if len(first) != 1:
        fail("%d baskets fit" % len(first))
    return match(options, first[0])
