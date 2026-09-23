def solve(options):
    numbers = set([int(a + b) for a, b in product(["1", "2"], repeat=2)])
    return match(options, len(numbers))
