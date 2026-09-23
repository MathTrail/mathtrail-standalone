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
    return match(options, month("Monday", 30)[15 - 1])
