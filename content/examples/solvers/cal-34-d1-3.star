def solve(options):
    return match(options, sum([days_in_month(2023, month_number) for month_number in [3, 4, 5]]))
