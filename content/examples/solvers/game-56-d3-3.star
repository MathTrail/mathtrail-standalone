def solve(options):
    sweets = 12
    eats = [1, 2]
    # wins[n] says whether the player about to move with n sweets left can
    # force a win; with none left, the other player ate the last one and lost.
    wins = [True]
    for n in range(1, sweets + 1):
        wins.append(any([not wins[n - eat] for eat in eats if eat <= n]))
    first = [eat for eat in eats if eat <= sweets and not wins[sweets - eat]]
    if len(first) != 1:
        fail("%d first moves win, and the question asks for the one" % len(first))
    return match(options, first[0])
