TRAIN_A = ("22:35", "6:10")  # leaves, arrives
TRAIN_B = ("23:50", "7:05")

def at(clock_time):
    parts = clock_time.split(":")
    return int(parts[0]) * 60 + int(parts[1])

def span(start, end):
    # Minutes from start forward to end, across midnight if need be.
    day = 24 * 60
    passed = 0
    now = start
    while now % day != end % day:
        now += 1
        passed += 1
    return passed

def counted(number, unit):
    # "1 minute", "2 minutes": a count as the options write it.
    return "%d %s" % (number, unit if number == 1 else unit + "s")

def solve(options):
    a = span(at(TRAIN_A[0]), at(TRAIN_A[1]))
    b = span(at(TRAIN_B[0]), at(TRAIN_B[1]))
    if a == b:
        return match(options, "They take the same time")
    longer = "Train A" if a > b else "Train B"
    return match(options, "%s, by %s" % (longer, counted(abs(a - b), "minute")))
