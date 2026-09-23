def one_pan(weights):
    # The loads some of the weights balance from the opposite pan.
    loads = set()
    for size in range(1, len(weights) + 1):
        for chosen in combinations(weights, size):
            loads.add(sum(chosen))
    return loads

def solve(options):
    loads = one_pan((1, 3))
    fitting = ["%d kg" % mass for mass in [1, 2, 3, 4] if mass not in loads]
    if len(fitting) != 1:
        fail("%d of the loads cannot be weighed" % len(fitting))
    return match(options, fitting[0])
