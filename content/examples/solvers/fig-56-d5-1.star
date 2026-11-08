def solve(options, picture):
    edges = set()
    for x1, y1, x2, y2 in picture["lines"]:
        if y1 == y2:
            for x in range(min(x1, x2), max(x1, x2)):
                edges.add((x, y1, "h"))
        elif x1 == x2:
            for y in range(min(y1, y2), max(y1, y2)):
                edges.add((x1, y, "v"))
    width = picture["width"]
    height = picture["height"]
    full = 0
    for y in range(height + 1):
        for x in range(width):
            full += 1
    for x in range(width + 1):
        for y in range(height):
            full += 1
    return match(options, full - len(edges))
