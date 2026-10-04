# Grandma's plane takes off at 8:00 by the clocks of her city and lands at
# 8:30 by the clocks of Ann's, which show 4 hours less: turn the landing into a
# time on the clocks of Grandma's city, count the minutes from take-off forward
# to it with span, and write them as hours and minutes.
# Written from the template minutes of time.clocks.

TAKES_OFF = "8:00"  # by the clocks of Grandma's city
LANDS = "8:30"  # by the clocks of Ann's city
BEHIND = 4 * 60  # minutes the clocks of Ann's city are behind Grandma's

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

def solve(options):
    flight = span(at(TAKES_OFF), at(LANDS) + BEHIND)
    return match(options, "%d hours %d minutes" % (flight // 60, flight % 60))
