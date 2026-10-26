def solve(options):
    # Prices in parts of a coin worth 1000, so that every price stays whole.
    coin = 1000
    answers = set()
    for milk in range(0, coin + 1):
        bread = coin - milk
        if bread % 2 != 0:
            continue
        if (bread // 2 + milk) * 125 != coin * 100:
            continue
        twice = milk * 125 * 120  # a rise of 25% and then one of 20%
        if twice % (coin * 100) != 0:
            continue
        answers.add(twice // (coin * 100))
    if len(answers) != 1:
        fail("the prices give %d answers" % len(answers))
    return match(options, "%d%%" % list(answers)[0])
