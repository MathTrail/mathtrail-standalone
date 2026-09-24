def solve(options):
    first = []
    for coins in range(1, 1001):
        if coins % 2 != 0:
            continue
        after_book = coins - coins // 2
        if after_book % 4 != 0:
            continue
        after_pen = after_book - after_book // 4  # a quarter of what was left
        if after_pen == 30:
            first.append(coins)
    if len(first) != 1:
        fail("%d starting amounts end with 30 coins" % len(first))
    return match(options, first[0])
