def solve(options):
    found = [n for n in range(100, 1000) if n // 100 + (n // 10) % 10 + n % 10 == 3]
    return match(options, len(found))
