STARTS = "22:40"
ENDS = "7:10"
BREAKS = 2
BREAK = 25  # minutes each

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
    # "1 hour", "2 hours": a count as the options write it.
    return "%d %s" % (number, unit if number == 1 else unit + "s")

def lasting(minutes):
    # A length of time as the options write it.
    return counted(minutes // 60, "hour") + " " + counted(minutes % 60, "minute")

def solve(options):
    return match(options, lasting(span(at(STARTS), at(ENDS)) - BREAKS * BREAK))
