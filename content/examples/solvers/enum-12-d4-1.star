def solve(options):
    numbers = set([int("".join(digits)) for digits in permutations(["1", "2", "3"])])
    return match(options, len(numbers))
