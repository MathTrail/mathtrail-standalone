def solve(options):
    marbles = 15
    takes = [1, 3, 4]
    wins = [False]  # with no marbles left, the other player took the last one
    for n in range(1, marbles + 1):
        wins.append(any([not wins[n - take] for take in takes if take <= n]))
    first = [take for take in takes if take <= marbles and not wins[marbles - take]]
    if len(first) == 0:
        return match(options, "She cannot win for certain")
    if len(first) > 1:
        fail("%d first moves win, and the question asks for the one" % len(first))
    return match(options, first[0])
