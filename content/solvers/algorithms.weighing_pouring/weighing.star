# For the fewest weighings on a balance without weights: fill in the answer for
# every number of coins from the bottom up, because a function may not call
# itself here. Putting k coins on each pan leaves the fake among the k on the
# pan that rises, or among the rest when the pans balance.
# From reference task wp-34-d4-1.

COINS = 12  # coins, one of them a lighter fake

def solve(options):
    best = {0: 0, 1: 0}  # weighings that find the fake among that many coins
    for count in range(2, COINS + 1):
        best[count] = min([1 + max([best[k], best[count - 2 * k]]) for k in range(1, count // 2 + 1)])
    return match(options, best[COINS])
