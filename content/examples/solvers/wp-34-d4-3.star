def one_pan(weights):
    # The loads some of the weights balance from the opposite pan.
    loads = set()
    for size in range(1, len(weights) + 1):
        for chosen in combinations(weights, size):
            loads.add(sum(chosen))
    return loads

def fewest_weights(loads, reach):
    for size in range(1, len(loads) + 1):
        for weights in combinations_with_replacement(range(1, max(loads) + 1), size):
            covered = reach(weights)
            if all([mass in covered for mass in loads]):
                return size
    fail("no set of weights covers every load")

def solve(options):
    return match(options, fewest_weights(range(1, 8), one_pan))
