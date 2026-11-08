def solve(options, picture):
    points = set([point["at"] for point in picture["points"]])
    found = set()
    for a in points:
        for b in points:
            if a == b:
                continue
            dx, dy = b[0] - a[0], b[1] - a[1]
            c = (b[0] - dy, b[1] + dx)
            d = (a[0] - dy, a[1] + dx)
            if c in points and d in points:
                found.add(tuple(sorted([a, b, c, d])))
    return match(options, len(found))
