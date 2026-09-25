DAYS = ["Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday", "Sunday"]
DAY = "Tuesday"
CALL_STARTS = "20:30"  # Moscow time
TALK = 45  # minutes
AHEAD = 7 * 60  # Vladivostok's clocks show this many minutes more than Moscow's

def at(clock_time):
    parts = clock_time.split(":")
    return int(parts[0]) * 60 + int(parts[1])

def show(total):
    # A time as the options write it: hours, then minutes padded by hand,
    # because % takes no width here.
    total = total % (24 * 60)
    minutes = total % 60
    return "%d:%s" % (total // 60, str(minutes) if minutes >= 10 else "0" + str(minutes))

def solve(options):
    # Minutes from the start of the week, on Vladivostok's clocks, when the call ends.
    ends = DAYS.index(DAY) * 24 * 60 + at(CALL_STARTS) + TALK + AHEAD
    return match(options, show(ends) + " on " + DAYS[(ends // (24 * 60)) % 7])
