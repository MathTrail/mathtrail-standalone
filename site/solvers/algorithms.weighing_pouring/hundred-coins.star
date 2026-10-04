# The fewest weighings on a balance without weights that find one lighter fake
# among 100 coins: fill in the answer for every number of coins from the
# bottom up. Putting k coins on each pan leaves the fake among the k on the
# pan that rises, or among the rest when the pans balance. The steps narrow
# the coins down to at most 34, 12, 4, 2 and 1, so each of those is checked
# to take one weighing fewer than the number before it.
# Written from the template weighing of algorithms.weighing_pouring.

COINS = 100  # coins, one of them a lighter fake
NARROWED = [100, 34, 12, 4, 2, 1]  # the most coins left after each weighing of the steps

def solve(options):
    best = {0: 0, 1: 0}  # weighings that find the fake among that many coins
    for count in range(2, COINS + 1):
        best[count] = min([1 + max([best[k], best[count - 2 * k]]) for k in range(1, count // 2 + 1)])
    for at in range(len(NARROWED) - 1):
        if best[NARROWED[at]] != best[NARROWED[at + 1]] + 1:
            fail("%d coins do not take one weighing more than %d" % (NARROWED[at], NARROWED[at + 1]))
    return match(options, best[COINS])
