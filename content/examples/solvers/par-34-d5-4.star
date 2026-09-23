def solve(options):
    results = set()
    for order in permutations(range(1, 7)):
        results.add(sum([order[i] if i % 2 == 0 else -order[i] for i in range(6)]))
    fitting = [n for n in [0, 2, 4, 5, 8] if n in results]
    if len(fitting) != 1:
        fail("%d of the numbers fit" % len(fitting))
    return match(options, fitting[0])
