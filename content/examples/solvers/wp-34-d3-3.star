def both_pans(weights):
    # The loads the weights balance when each is opposite the load, beside it,
    # or left off.
    loads = set()
    for signs in product([1, -1, 0], repeat=len(weights)):
        total = sum([signs[i] * weights[i] for i in range(len(weights))])
        if total > 0:
            loads.add(total)
    return loads

def solve(options):
    loads = both_pans((1, 3, 9))
    fitting = ["%d kg" % mass for mass in [2, 5, 7, 11, 14] if mass not in loads]
    if len(fitting) != 1:
        fail("%d of the loads cannot be weighed" % len(fitting))
    return match(options, fitting[0])
