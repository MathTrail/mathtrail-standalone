def solve(options):
    return match(options, len([year for year in range(2017, 2032 + 1) if is_leap(year)]))
