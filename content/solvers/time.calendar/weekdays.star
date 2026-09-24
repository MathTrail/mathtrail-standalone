# For the day of the week some days away: count the days between the two
# dates with days_between, then step through the week that many days. A task
# set "in some year" still needs a real year for its dates: choose one in
# which the weekday the task gives is true, and check that it is.
# From reference task cal-34-d5-3.

# The days of the week as the options write them, Monday first, which is how
# weekday() counts them.
WEEK = ["Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday", "Sunday"]
YEAR = 2026  # a year in which 10 March is a Tuesday
KNOWN = (YEAR, 3, 10)  # the date whose weekday the task gives
KNOWN_DAY = "Tuesday"
ASKED = (YEAR, 5, 10)  # the date the question asks about

def shift(day, days):
    # One day at a time, forward or back.
    index = WEEK.index(day)
    step = 1 if days >= 0 else -1
    for _ in range(abs(days)):
        index = (index + step) % 7
    return WEEK[index]

def solve(options):
    if WEEK[weekday(KNOWN[0], KNOWN[1], KNOWN[2])] != KNOWN_DAY:
        fail("in %d that date is not a %s" % (YEAR, KNOWN_DAY))
    return match(options, shift(KNOWN_DAY, days_between(KNOWN, ASKED)))
