def solve(options):
    opened = [number for number in range(1, 21) if number % 2 == 0]
    return match(options, len([number for number in range(1, 21) if number not in opened]))
