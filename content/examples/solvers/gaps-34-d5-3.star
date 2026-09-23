def solve(options):
    marks = range(0, 20 + 1, 4)
    border = set()
    for x in marks:
        for y in marks:
            if x == 0 or x == 20 or y == 0 or y == 20:
                border.add((x, y))
    return match(options, len(border))
