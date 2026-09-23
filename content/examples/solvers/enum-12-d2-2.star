def solve(options):
    numbers = set([int(a + b) for a, b in permutations(["3", "5", "7"], 2)])
    return match(options, len(numbers))
