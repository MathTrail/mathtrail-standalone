def signed_results(last):
    # Every way of putting a plus or a minus before each of 2, 3, ..., last.
    results = set()
    for signs in product([1, -1], repeat=last - 1):
        results.add(1 + sum([signs[i] * (i + 2) for i in range(last - 1)]))
    return results

def solve(options):
    results = signed_results(9)
    fitting = [n for n in [0, 2, 3, 4, 6] if n in results]
    if len(fitting) != 1:
        fail("%d of the numbers fit" % len(fitting))
    return match(options, fitting[0])
