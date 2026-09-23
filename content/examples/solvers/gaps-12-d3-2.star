def ring_gaps(objects):
    edges = set()
    for i in range(objects):
        j = (i + 1) % objects
        edges.add((min([i, j]), max([i, j])))
    return len(edges)

def solve(options):
    return match(options, ring_gaps(8))
