# The days of the week as the options write them, Monday first, which is how
# weekday() counts them.
WEEK = ["Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday", "Sunday"]
KNOWN = (2024, 3, 4)  # the date whose weekday the task gives: Anya's birth
KNOWN_DAY = "Monday"
ASKED = (2034, 3, 4)  # her 10th birthday

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
    return match(options, shift(KNOWN_DAY, days_between(KNOWN, ASKED)))
