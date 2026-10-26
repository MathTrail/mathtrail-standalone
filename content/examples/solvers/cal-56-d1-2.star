# The days of the week as the options write them, Monday first, which is how
# weekday() counts them.
WEEK = ["Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday", "Sunday"]
KNOWN = (2026, 2, 13)  # the date whose weekday the task gives
KNOWN_DAY = "Friday"
ASKED = (2026, 3, 13)
FEBRUARY = 28  # the days the task says February 2026 has

def shift(day, days):
    # One day at a time, forward or back.
    index = WEEK.index(day)
    step = 1 if days >= 0 else -1
    for _ in range(abs(days)):
        index = (index + step) % 7
    return WEEK[index]

def solve(options):
    if WEEK[weekday(KNOWN[0], KNOWN[1], KNOWN[2])] != KNOWN_DAY:
        fail("that date is not a %s" % KNOWN_DAY)
    if days_in_month(2026, 2) != FEBRUARY:
        fail("February 2026 does not have %d days" % FEBRUARY)
    return match(options, shift(KNOWN_DAY, days_between(KNOWN, ASKED)))
