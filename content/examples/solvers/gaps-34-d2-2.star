def gaps(objects):
    return len([(i, i + 1) for i in range(objects - 1)])

def solve(options):
    per_gap = 6 // gaps(3)
    return match(options, gaps(7) * per_gap)
