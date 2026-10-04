# Kate waters the flowers every 3 days, the first time on a Monday: count the
# gaps between neighbouring waterings up to the 4th, step through the week
# that many days, and read the day. The steps say the days are a week and 2
# days more, so that is checked too.
# Written from the template weekdays of time.calendar.

# The days of the week as the answer writes them, Monday first.
WEEK = ["Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday", "Sunday"]
FIRST = "Monday"  # the day of the first watering
EVERY = 3  # days from one watering to the next
ASKED = 4  # the watering the question asks about

def shift(day, days):
    # One day at a time, forward or back.
    index = WEEK.index(day)
    step = 1 if days >= 0 else -1
    for _ in range(abs(days)):
        index = (index + step) % 7
    return WEEK[index]

def solve(options):
    gaps = len([(watering, watering + 1) for watering in range(1, ASKED)])
    days = gaps * EVERY
    if days != 7 + 2:
        fail("%d days are not a week and 2 days" % days)
    return match(options, shift(FIRST, days))
