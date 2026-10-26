def solve(options):
    # Grass in cow-days: what the meadow holds at the start and what grows a
    # day, tried for every whole growth; the first herd then gives the start.
    cows = set()
    for growth in range(0, 101):
        start = 24 * 10 - growth * 10
        if start < 0 or 14 * 20 != start + growth * 20:
            continue
        if (start + growth * 40) % 40 == 0:
            cows.add((start + growth * 40) // 40)
    if len(cows) != 1:
        fail("%d herds fit" % len(cows))
    return match(options, list(cows)[0])
