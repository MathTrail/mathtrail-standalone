def at(clock_time):
    parts = clock_time.split(":")
    return int(parts[0]) * 60 + int(parts[1])

def oclock(total):
    if total % 60 != 0:
        fail("that is not a whole hour")
    hour = (total // 60) % 12
    if hour == 0:
        hour = 12
    return "%d o'clock" % hour

def solve(options):
    return match(options, oclock(at("12:00") - 2 * 60))
