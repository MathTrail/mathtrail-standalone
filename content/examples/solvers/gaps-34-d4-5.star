def solve(options):
    lamps = set()
    for i in range(11):
        lamps.add(30 * i)
    for i in range(10):
        lamps.add(30 * i + 15)
    return match(options, len(lamps))
