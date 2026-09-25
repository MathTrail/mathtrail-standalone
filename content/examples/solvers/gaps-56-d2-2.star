WIDTH = 20  # metres
LENGTH = 30  # metres
APART = 5  # metres between rows and between trees in a row

def solve(options):
    trees = set()
    for x in range(0, WIDTH + 1, APART):
        for y in range(0, LENGTH + 1, APART):
            trees.add((x, y))  # the edges and the corners are included
    return match(options, len(trees))
