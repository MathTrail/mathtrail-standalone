def solve(options):
    # Every whole price the toy could have now and was planned to have: 20%
    # below and 20% above its old price of 50 coins.
    old = 50
    gaps = set()
    for now in range(0, 2 * old + 1):
        for planned in range(0, 2 * old + 1):
            if now * 100 == old * (100 - 20) and planned * 100 == old * (100 + 20):
                gaps.add(planned - now)
    if len(gaps) != 1:
        fail("%d gaps fit" % len(gaps))
    return match(options, list(gaps)[0])
