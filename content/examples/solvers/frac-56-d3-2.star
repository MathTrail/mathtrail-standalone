def solve(options):
    full = 1500  # jug and juice together
    juices = []
    for juice in range(1, full + 1):
        jug = full - juice
        if juice % 3 == 0 and jug + juice // 3 == 700:
            juices.append(juice)
    if len(juices) != 1:
        fail("%d amounts of juice fit both weights" % len(juices))
    return match(options, juices[0])
