def solve(options):
    numbers = set([int("".join(digits)) for digits in product(["1", "2", "3"], repeat=3)])
    return match(options, len(numbers))
