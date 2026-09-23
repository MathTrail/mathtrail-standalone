def at(clock_time):
    parts = clock_time.split(":")
    return int(parts[0]) * 60 + int(parts[1])

def show(total):
    total = total % (24 * 60)
    minutes = total % 60
    return "%d:%s" % (total // 60, str(minutes) if minutes >= 10 else "0" + str(minutes))

def solve(options):
    # The clock runs 63 minutes for every 60 real ones.
    shown = at("12:12") - at("8:00")
    for real in range(1, 1000):
        if real * 63 == shown * 60:
            return match(options, show(at("8:00") + real))
    fail("no real time fits what the clock shows")
