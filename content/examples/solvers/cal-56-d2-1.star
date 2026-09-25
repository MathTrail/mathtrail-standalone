# The days of the week as the options write them, Monday first, which is how
# weekday() counts them.
WEEK = ["Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday", "Sunday"]
KNOWN = (2027, 3, 1)  # the date whose weekday the task gives
KNOWN_DAY = "Monday"
ASKED = (2028, 3, 1)  # the date the question asks about
LEAP = 2028  # the year the task says is a leap year

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
    if not is_leap(LEAP):
        fail("%d is not a leap year" % LEAP)
    return match(options, shift(KNOWN_DAY, days_between(KNOWN, ASKED)))
