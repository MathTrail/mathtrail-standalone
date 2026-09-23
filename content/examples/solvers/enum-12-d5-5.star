def solve(options):
    numbers = [int(a + b) for a, b in product(["0", "1", "2"], repeat=2)]
    two_digit = set([n for n in numbers if n >= 10 and n <= 99])
    return match(options, len(two_digit))
