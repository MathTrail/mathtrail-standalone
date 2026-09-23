def solve(options):
    found = [n for n in range(10, 100) if n // 10 + n % 10 == 5]
    return match(options, len(found))
