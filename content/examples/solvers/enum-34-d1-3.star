def solve(options):
    numbers = set([int("".join(digits)) for digits in permutations(["4", "5", "6", "7"], 3)])
    return match(options, len(numbers))
