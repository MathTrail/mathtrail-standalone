def solve(options):
    takes = [1, 3]
    starts = [10, 11, 12, 13, 14]
    wins = [False]  # with no cards left, the other player removed the last one
    for n in range(1, max(starts) + 1):
        wins.append(any([not wins[n - take] for take in takes if take <= n]))
    return match(options, len([start for start in starts if wins[start]]))
