NAMES = ["Ann", "Ben", "Dan", "Eve", "Fay", "Kim"]
TALLER = [("Ann", "Dan"), ("Dan", "Ben"), ("Dan", "Eve"), ("Ben", "Kim"), ("Eve", "Kim"), ("Kim", "Fay")]

def solve(options):
    # Every order of the six by height that fits the results; two children
    # cannot be put in order when some fitting orders have them one way round
    # and some the other.
    seen = {}
    for order in permutations(NAMES):
        at = {name: i for i, name in enumerate(order)}
        if not all([at[a] < at[b] for a, b in TALLER]):
            continue
        for a, b in combinations(NAMES, 2):
            seen[(a, b, at[a] < at[b])] = True
    open_pairs = [(a, b) for a, b in combinations(NAMES, 2) if seen.get((a, b, True)) and seen.get((a, b, False))]
    if not open_pairs:
        return match(options, "None: all six can be put in order")
    if len(open_pairs) != 1:
        fail("%d pairs are open" % len(open_pairs))
    return match(options, open_pairs[0][0] + " and " + open_pairs[0][1])
