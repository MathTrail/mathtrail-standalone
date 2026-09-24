def solve(options):
    before = []
    for price in range(1, 1001):
        if price * 20 % 100 != 0:
            continue
        if price - price * 20 // 100 == 60:
            before.append(price)
    if len(before) != 1:
        fail("%d prices end at 60 after the discount" % len(before))
    return match(options, before[0])
