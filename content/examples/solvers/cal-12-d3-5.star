def solve(options):
    fitting = [year for year in range(2019, 2100) if year - 2019 == 10]
    if len(fitting) != 1:
        fail("%d years fit the clue" % len(fitting))
    return match(options, fitting[0])
