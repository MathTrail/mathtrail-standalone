def solve(options):
    numbers = set([int("".join(digits)) for digits in permutations(["0", "1", "2", "3"]) if digits[0] != "0"])
    return match(options, len(numbers))
