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

def show(total):
    total = total % (24 * 60)
    minutes = total % 60
    return "%d:%s" % (total // 60, str(minutes) if minutes >= 10 else "0" + str(minutes))

def solve(options):
    swift = span(at("7:50"), at("10:35"))
    return match(options, show(at("8:20") + swift + 25))
