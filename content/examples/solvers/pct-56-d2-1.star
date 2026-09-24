def solve(options):
    green = 6 + 5
    marbles = 6 + 14 + 5
    # The percentage p with green out of marbles equal to p out of 100.
    percents = [p for p in range(0, 101) if green * 100 == p * marbles]
    if len(percents) != 1:
        fail("%d out of %d is not a whole percentage" % (green, marbles))
    return match(options, "%d%%" % percents[0])
