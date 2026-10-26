def solve(options):
    # Every number of children and of sweets each, more than one sweet each.
    totals = set()
    for children in range(2, 101):
        for each in range(2, 101):
            if each + 5 == (children - 1) * each:
                totals.add(children * each)
    if len(totals) != 1:
        fail("%d totals fit" % len(totals))
    return match(options, list(totals)[0])
