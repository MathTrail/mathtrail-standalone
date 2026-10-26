def solve(options):
    # Every whole number of coins after three years that holds the first 500
    # and three yearly additions of 20% of those first 500 coins.
    first, percent, years = 500, 20, 3
    coins = [c for c in range(0, 10001) if (c - first) * 100 == years * percent * first]
    if len(coins) != 1:
        fail("%d sums fit" % len(coins))
    return match(options, coins[0])
