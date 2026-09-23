def gaps(objects):
    return len([(i, i + 1) for i in range(objects - 1)])

def solve(options):
    for flowers in range(1, 50):
        if gaps(flowers) == 4:
            return match(options, flowers)
    fail("no row of flowers has four gaps")
