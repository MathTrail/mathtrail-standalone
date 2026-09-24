ROUND = 360  # metres round the closed path
OLD = 15  # metres between the old lamps
NEW = 40  # metres between the new lamps

def solve(options):
    # Places are measured from the first new lamp, which stands at an old lamp's place.
    # Each range stops before ROUND: on a closed path that place is the place at 0.
    old = set(range(0, ROUND, OLD))
    new = set(range(0, ROUND, NEW))
    return match(options, len([place for place in new if place not in old]))  # new holes
