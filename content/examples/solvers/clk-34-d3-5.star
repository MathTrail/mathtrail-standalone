def at(clock_time):
    parts = clock_time.split(":")
    return int(parts[0]) * 60 + int(parts[1])

def span(start, end):
    day = 24 * 60
    passed = 0
    now = start
    while now % day != end % day:
        now += 1
        passed += 1
    return passed

def solve(options):
    total = span(at("9:15"), at("11:45"))
    for lesson in range(1, 200):
        if 3 * lesson + 2 * 15 == total:
            return match(options, lesson)
    fail("no lesson length fits the morning")
