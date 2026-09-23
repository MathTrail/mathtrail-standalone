def solve(options):
    return match(options, len([start for start in range(0, 366, 7) if start + 7 <= 366]))
