# The days of the week as the options write them, Monday first, which is how
# weekday() counts them.
WEEK = ["Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday", "Sunday"]
MONTH, DAY = 6, 5  # 5 June
KNOWN_YEAR = 2026
KNOWN_DAY = "Friday"

def solve(options):
    if WEEK[weekday(KNOWN_YEAR, MONTH, DAY)] != KNOWN_DAY:
        fail("that date is not a %s" % KNOWN_DAY)
    # Go on a year at a time until the date falls on the same day again.
    year = KNOWN_YEAR + 1
    while WEEK[weekday(year, MONTH, DAY)] != KNOWN_DAY:
        year += 1
    return match(options, year)
