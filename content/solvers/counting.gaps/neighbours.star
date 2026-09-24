# For something paid for gap by gap, such as steps per flight of stairs or
# seconds between rings: count the gaps as pairs of neighbours, work out what
# one gap is worth from what the task tells, and carry it over to what it asks.
# From reference task gaps-34-d4-4.

TOLD_FROM = 1  # the floor the counting starts at
TOLD_TO = 3  # the floor it stops at
TOLD_STEPS = 36  # the steps counted between them
ASKED_FROM = 3  # the floors the question asks about
ASKED_TO = 7

def flights(start, end):
    # One flight between every floor and the next.
    return len([(floor, floor + 1) for floor in range(start, end)])

def solve(options):
    told = flights(TOLD_FROM, TOLD_TO)
    if TOLD_STEPS % told != 0:
        fail("%d steps do not share out evenly over %d flights" % (TOLD_STEPS, told))
    return match(options, flights(ASKED_FROM, ASKED_TO) * (TOLD_STEPS // told))
