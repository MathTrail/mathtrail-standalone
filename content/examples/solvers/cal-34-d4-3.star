WEEK = ["Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday", "Sunday"]

def shift(day, days):
    # One day at a time, forward or back.
    index = WEEK.index(day)
    step = 1 if days >= 0 else -1
    for _ in range(abs(days)):
        index = (index + step) % 7
    return WEEK[index]

def month(first_day, length):
    return [shift(first_day, n) for n in range(length)]

def solve(options):
    firsts = set()
    for first in WEEK:
        for length in range(28, 32):
            days = month(first, length)
            count = {day: len([d for d in days if d == day]) for day in WEEK}
            if (count["Friday"] == 5 and count["Saturday"] == 5 and
                    count["Thursday"] == 4 and count["Sunday"] == 4):
                firsts.add(first)
    if len(firsts) != 1:
        fail("%d first days fit the clue" % len(firsts))
    return match(options, list(firsts)[0])
