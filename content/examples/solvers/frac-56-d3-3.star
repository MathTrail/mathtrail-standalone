def solve(options):
    totals = []
    for cars in range(1, 1001):
        if cars % 8 != 0:
            continue
        red = cars // 8 * 5
        blue = cars - red
        if red - blue == 6:
            totals.append(cars)
    if len(totals) != 1:
        fail("%d car parks fit" % len(totals))
    return match(options, totals[0])
