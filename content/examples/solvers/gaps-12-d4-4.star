def solve(options):
    side = range(3)
    border = set()
    for x in side:
        for y in side:
            if x == 0 or x == 2 or y == 0 or y == 2:
                border.add((x, y))
    return match(options, len(border))
