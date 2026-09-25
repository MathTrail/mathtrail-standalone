STONES = 4

def splits(possible):
    # For each pair of stones, the orders still possible in which the first is
    # heavier and those in which the second is. An order gives each stone its
    # place, 0 for the heaviest.
    parts = []
    for first, second in combinations(range(STONES), 2):
        heavier = tuple([order for order in possible if order[first] < order[second]])
        lighter = tuple([order for order in possible if order[first] > order[second]])
        if heavier and lighter:
            parts.append((heavier, lighter))
    return parts

def solve(options):
    start = tuple(permutations(range(STONES)))
    # Every set of orders the weighings can leave, found breadth first.
    parts = {}
    states = [start]
    seen = set([start])
    head = 0
    while head < len(states):
        state = states[head]
        head += 1
        parts[state] = splits(state)
        for pair in parts[state]:
            for part in pair:
                if part not in seen:
                    seen.add(part)
                    states.append(part)
    # The fewest weighings for each, smaller sets first: a weighing always
    # leaves a smaller set, so its parts are known by then. The order is known
    # when a single order is left.
    fewest = {}
    for state in sorted(states, key=len):
        if len(state) == 1:
            fewest[state] = 0
        else:
            fewest[state] = min([1 + max([fewest[heavier], fewest[lighter]]) for heavier, lighter in parts[state]])
    return match(options, fewest[start])
