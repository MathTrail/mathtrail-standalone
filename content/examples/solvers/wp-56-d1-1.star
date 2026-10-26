COINS = 10  # one of them a heavier fake

def solve(options):
    # k coins on each pan leave the fake among the k on the pan that sinks, or
    # among the rest when the pans balance; the fewest weighings for every
    # number of coins are filled in from the bottom up, because a function may
    # not call itself here.
    best = {0: 0, 1: 0}
    for count in range(2, COINS + 1):
        best[count] = min([1 + max([best[k], best[count - 2 * k]]) for k in range(1, count // 2 + 1)])
    return match(options, best[COINS])
