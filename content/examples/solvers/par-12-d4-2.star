def solve(options):
    return match(options, len([number for number in range(1, 16) if number % 2 == 1]))
