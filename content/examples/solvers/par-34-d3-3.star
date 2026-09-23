def solve(options):
    ann = [number for number in range(1, 101) if number % 2 == 1]
    return match(options, len([number for number in ann if number % 10 != 5]))
