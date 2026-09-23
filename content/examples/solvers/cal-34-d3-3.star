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
    most = 0
    for first in WEEK:
        for length in range(28, 32):
            days = month(first, length)
            most = max([most, len([day for day in days if day == "Saturday"])])
    return match(options, most)
