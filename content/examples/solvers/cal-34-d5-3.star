WEEK = ["Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday", "Sunday"]

def shift(day, days):
    # One day at a time, forward or back.
    index = WEEK.index(day)
    step = 1 if days >= 0 else -1
    for _ in range(abs(days)):
        index = (index + step) % 7
    return WEEK[index]

def solve(options):
    if WEEK[weekday(2026, 3, 10)] != "Tuesday":
        fail("10 March 2026 is not a Tuesday")
    return match(options, shift("Tuesday", days_between((2026, 3, 10), (2026, 5, 10))))
