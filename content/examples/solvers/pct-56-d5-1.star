def solve(options):
    # Every first go of whole wheels and bolts that reaches the 320 coins; all
    # that is bought after it costs 80% of its price.
    wheel, bolt, need_wheels, need_bolts = 150, 10, 3, 1
    best = need_wheels * wheel + need_bolts * bolt  # everything at full price
    for wheels in range(0, need_wheels + 1):
        for bolts in range(0, 41):
            paid = wheels * wheel + bolts * bolt
            if paid < 320:
                continue
            rest = (need_wheels - wheels) * wheel + max(0, need_bolts - bolts) * bolt
            if rest * 80 % 100 != 0:
                fail("a discounted price is not a whole number of coins")
            best = min(best, paid + rest * 80 // 100)
    return match(options, best)
