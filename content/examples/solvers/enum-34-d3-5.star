def solve(options):
    children = ["b1", "b2", "b3", "g1", "g2", "g3", "g4"]
    mixed = []
    for pair in combinations(children, 2):
        if not (pair[0].startswith("g") and pair[1].startswith("g")):
            mixed.append(pair)
    return match(options, len(mixed))
