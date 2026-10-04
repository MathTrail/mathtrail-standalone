# In some month three Saturdays fall on even dates: try every day of the week
# for the 1st and every length a month can have, keep the months in which
# exactly three Saturdays are even, and read the day of the 25th. Exactly one
# day must be left. The steps say those Saturdays are the 2nd, the 16th and
# the 30th, so that is checked too.
# Written from the template weekdays of time.calendar.

# The days of the week as the answer writes them, Monday first.
WEEK = ["Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday", "Sunday"]
ASKED = 25  # the date the question asks about

def shift(day, days):
    # One day at a time, forward or back.
    index = WEEK.index(day)
    step = 1 if days >= 0 else -1
    for _ in range(abs(days)):
        index = (index + step) % 7
    return WEEK[index]

def solve(options):
    answers = set()
    for first in WEEK:
        for length in [28, 29, 30, 31]:
            saturdays = [date for date in range(1, length + 1) if shift(first, date - 1) == "Saturday"]
            even = [date for date in saturdays if date % 2 == 0]
            if len(even) == 3:
                if even != [2, 16, 30]:
                    fail("the even Saturdays are not the 2nd, the 16th and the 30th")
                answers.add(shift(first, ASKED - 1))
    if len(answers) != 1:
        fail("the month fits %d days for the %dth" % (len(answers), ASKED))
    return match(options, list(answers)[0])
