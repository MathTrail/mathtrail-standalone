# The days of the week as weekday() counts them, Monday first.
WEEK = ["Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday", "Sunday"]
SCHOOL = ["Sunday", "Monday", "Tuesday", "Wednesday", "Thursday"]
YEAR = 2029  # a year in which 1 April is a Sunday

def solve(options):
    if WEEK[weekday(YEAR, 4, 1)] != "Sunday":
        fail("in %d 1 April is not a Sunday" % YEAR)
    days = range(1, days_in_month(YEAR, 4) + 1)
    return match(options, len([day for day in days if WEEK[weekday(YEAR, 4, day)] in SCHOOL]))
