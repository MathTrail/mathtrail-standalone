# 12 matches laid as a square of 2 by 2 cells: try every set of matches to take
# away, the smallest first, and stop at the first that leaves no square of any
# size whole.
# Written from the template squares-on-lines of counting.figures.

WIDTH = 2  # cells across
HEIGHT = 2  # cells down

def whole(edges, x, y, size):
    for k in range(size):
        if (x + k, y, "h") not in edges or (x + k, y + size, "h") not in edges:
            return False
        if (x, y + k, "v") not in edges or (x + size, y + k, "v") not in edges:
            return False
    return True

def squares(edges):
    count = 0
    for size in range(1, min(WIDTH, HEIGHT) + 1):
        for x in range(WIDTH - size + 1):
            for y in range(HEIGHT - size + 1):
                if whole(edges, x, y, size):
                    count += 1
    return count

def solve(options):
    edges = [(x, y, "h") for y in range(HEIGHT + 1) for x in range(WIDTH)]
    edges += [(x, y, "v") for x in range(WIDTH + 1) for y in range(HEIGHT)]
    for k in range(len(edges) + 1):
        for taken in combinations(edges, k):
            left = set(edges)
            for edge in taken:
                left.remove(edge)
            if squares(left) == 0:
                return match(options, k)
    return []
