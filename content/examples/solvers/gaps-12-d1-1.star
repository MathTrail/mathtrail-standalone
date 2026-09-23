def gaps(objects):
    return len([(i, i + 1) for i in range(objects - 1)])

def solve(options):
    return match(options, gaps(6))
