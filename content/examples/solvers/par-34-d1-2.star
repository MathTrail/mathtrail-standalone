def solve(options):
    return match(options, len([number for number in range(1, 100) if number % 2 == 1]))
