SHOPS = 3
MORE = 2  # coins spent in each shop after the half

def left_after_shops(coins):
    # None where a half would not be whole coins or the coins would run out.
    for _ in range(SHOPS):
        if coins % 2 != 0:
            return None
        coins = coins // 2 - MORE
        if coins < 0:
            return None
    return coins

def solve(options):
    fitting = [n for n in range(1001) if left_after_shops(n) == 0]
    if len(fitting) != 1:
        fail("%d numbers of coins fit the story" % len(fitting))
    return match(options, fitting[0])
