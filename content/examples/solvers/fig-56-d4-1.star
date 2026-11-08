def unit_edges(lines):
    # Every line along the paper's lines, split into the sides of cells it covers.
    edges = set()
    for x1, y1, x2, y2 in lines:
        if y1 == y2:
            for x in range(min(x1, x2), max(x1, x2)):
                edges.add((x, y1, "h"))
        elif x1 == x2:
            for y in range(min(y1, y2), max(y1, y2)):
                edges.add((x1, y, "v"))
    return edges

def whole(edges, x, y, size):
    # Whether the square of this size at corner (x, y) has every side drawn.
    for k in range(size):
        if (x + k, y, "h") not in edges or (x + k, y + size, "h") not in edges:
            return False
        if (x, y + k, "v") not in edges or (x + size, y + k, "v") not in edges:
            return False
    return True

def squares(edges, width, height):
    # The squares of every size whose four sides are all drawn.
    found = []
    for size in range(1, min(width, height) + 1):
        for x in range(width - size + 1):
            for y in range(height - size + 1):
                if whole(edges, x, y, size):
                    found.append((x, y, size))
    return found

def solve(options, picture):
    edges = sorted(unit_edges(picture["lines"]))
    for k in range(len(edges) + 1):
        for taken in combinations(edges, k):
            left = set(edges)
            for edge in taken:
                left.remove(edge)
            if not squares(left, picture["width"], picture["height"]):
                return match(options, k)
    return []
