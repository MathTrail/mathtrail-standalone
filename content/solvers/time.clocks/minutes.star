# For clock times and how long things take: turn every time into minutes since
# midnight, count forward with span so that a trip past midnight comes out
# right, and turn the answer back into a time with show.
# From reference task clk-34-d5-3.

SWIFT_LEAVES = "7:50"  # the first train's trip
SWIFT_ARRIVES = "10:35"
PONY_LEAVES = "8:20"  # when the second train leaves
PONY_SLOWER = 25  # minutes its trip takes longer

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

def show(total):
    # A time as the options write it: hours, then minutes padded by hand,
    # because % takes no width here.
    total = total % (24 * 60)
    minutes = total % 60
    return "%d:%s" % (total // 60, str(minutes) if minutes >= 10 else "0" + str(minutes))

def solve(options):
    trip = span(at(SWIFT_LEAVES), at(SWIFT_ARRIVES))
    return match(options, show(at(PONY_LEAVES) + trip + PONY_SLOWER))
