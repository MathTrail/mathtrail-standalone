# Flags hang in a pattern that repeats: step along the row one flag at a time,
# round and round the pattern, instead of counting the groups by hand.
# Written from the template weekdays of time.calendar.

PATTERN = ["red", "yellow", "green", "blue"]  # the colours in order, as the options write them
ASKED = 23  # the flag the question asks about
WHOLE = 5  # fours before it, as the steps count them

def solve(options):
    if ASKED != len(PATTERN) * WHOLE + 3:
        fail("flag %d is not the third after %d fours" % (ASKED, WHOLE))
    index = 0
    for _ in range(ASKED - 1):
        index = (index + 1) % len(PATTERN)
    return match(options, PATTERN[index])
