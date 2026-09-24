SMALL = 60  # metres round the small track
BIG = 100  # metres round the big track
APART = 10  # metres between flags

def solve(options):
    # Places are measured round each closed track from the touching point, which is one place on both.
    flags = set()
    for place in range(0, SMALL, APART):
        flags.add(("small", place) if place != 0 else "touching point")
    for place in range(0, BIG, APART):
        flags.add(("big", place) if place != 0 else "touching point")
    return match(options, len(flags))
