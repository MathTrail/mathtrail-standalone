def solve(options):
    runs = [start for start in range(100) if sum(range(start, start + 5)) == 100]
    if len(runs) != 1:
        fail("%d runs of five numbers in a row add up to 100" % len(runs))
    return match(options, runs[0] + 4)
