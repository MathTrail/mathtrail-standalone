def light_weighings(coins):
    # Recursion is off, so the fewest weighings are filled in from the bottom:
    # k coins on each pan leave the fake among the k on the pan that rises, or
    # among the rest when they balance.
    best = {0: 0, 1: 0}
    for count in range(2, coins + 1):
        best[count] = min([1 + max([best[k], best[count - 2 * k]]) for k in range(1, count // 2 + 1)])
    return best[coins]

def solve(options):
    return match(options, light_weighings(3))
