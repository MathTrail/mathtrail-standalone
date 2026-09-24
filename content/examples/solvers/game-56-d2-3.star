def solve(options):
    matches = 12
    takes = [1, 2, 3]
    wins = [False]  # with no matches left, the other player took the last one
    for n in range(1, matches + 1):
        wins.append(any([not wins[n - take] for take in takes if take <= n]))
    if not wins[matches]:
        return match(options, "Max, whatever Zoe does")
    first = [take for take in takes if take <= matches and not wins[matches - take]]
    if len(first) == len(takes):
        return match(options, "Zoe, whatever she does")
    if len(first) > 1:
        fail("%d first moves win, and no option says which" % len(first))
    return match(options, "Zoe, if she picks up %d first" % first[0])
