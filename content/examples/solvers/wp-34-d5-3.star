def both_pans(weights):
    # The loads the weights balance when each is opposite the load, beside it,
    # or left off.
    loads = set()
    for signs in product([1, -1, 0], repeat=len(weights)):
        total = sum([signs[i] * weights[i] for i in range(len(weights))])
        if total > 0:
            loads.add(total)
    return loads

def fewest_weights(loads, reach):
    for size in range(1, len(loads) + 1):
        for weights in combinations_with_replacement(range(1, max(loads) + 1), size):
            covered = reach(weights)
            if all([mass in covered for mass in loads]):
                return size
    fail("no set of weights covers every load")

def solve(options):
    return match(options, fewest_weights(range(1, 14), both_pans))
