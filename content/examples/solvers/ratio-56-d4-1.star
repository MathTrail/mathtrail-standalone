def solve(options):
    daisies = 30
    # One link of the chain at a time: 4 tulips for every 5 daisies, then 2 roses for every 3 tulips.
    tulips = [t for t in range(0, 1001) if t * 5 == daisies * 4]
    if len(tulips) != 1:
        fail("%d numbers of tulips fit" % len(tulips))
    roses = [r for r in range(0, 1001) if r * 3 == tulips[0] * 2]
    if len(roses) != 1:
        fail("%d numbers of roses fit" % len(roses))
    return match(options, daisies + tulips[0] + roses[0])
