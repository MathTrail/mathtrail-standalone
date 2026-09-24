def solve(options):
    stones = 10
    takes = [1, 2]
    # wins[n] says whether the player about to move with n stones left can
    # force a win; with none left, the other player took the last one.
    wins = [False]
    for n in range(1, stones + 1):
        wins.append(any([not wins[n - take] for take in takes if take <= n]))
    first = [take for take in takes if take <= stones and not wins[stones - take]]
    if len(first) == 0:
        return match(options, "She cannot be sure to win")
    if len(first) > 1:
        fail("%d first moves win, and the question asks for the one" % len(first))
    return match(options, first[0])
