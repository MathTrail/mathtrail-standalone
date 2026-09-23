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

def duration(total):
    hours = total // 60
    minutes = total % 60
    if hours == 0:
        return "%d minutes" % minutes
    spoken = "%d hour" % hours
    if hours > 1:
        spoken += "s"
    if minutes == 0:
        return spoken
    return "%s %d minutes" % (spoken, minutes)

def solve(options):
    return match(options, duration(span(at("22:40"), at("1:15"))))
