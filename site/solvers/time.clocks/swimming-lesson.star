# A swimming lesson starts at 3:45 and lasts 40 minutes: turn the start into
# minutes since midnight, add the lesson, and turn the end back into a time
# with show. The steps go through 4:00, 15 minutes after the start, so that is
# checked too.
# Written from the template minutes of time.clocks.

STARTS = "3:45"
LASTS = 40  # minutes
WHOLE_HOUR = "4:00"  # the whole hour the steps go through

def at(clock_time):
    parts = clock_time.split(":")
    return int(parts[0]) * 60 + int(parts[1])

def show(total):
    # A time as the answer writes it: hours, then minutes padded by hand,
    # because % takes no width here.
    total = total % (24 * 60)
    minutes = total % 60
    return "%d:%s" % (total // 60, str(minutes) if minutes >= 10 else "0" + str(minutes))

def solve(options):
    if at(WHOLE_HOUR) - at(STARTS) != 15:
        fail("from %s to %s is not 15 minutes" % (STARTS, WHOLE_HOUR))
    return match(options, show(at(STARTS) + LASTS))
