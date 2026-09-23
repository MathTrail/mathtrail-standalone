WEEK = ["Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday", "Sunday"]

def solve(options):
    first = (2028, 1, 1)
    if not is_leap(2028) or WEEK[weekday(first[0], first[1], first[2])] != "Saturday":
        fail("2028 is not a leap year starting on a Saturday")
    saturdays = 0
    for n in range(366):
        day = add_days(first, n)
        if WEEK[weekday(day[0], day[1], day[2])] == "Saturday":
            saturdays += 1
    return match(options, saturdays)
