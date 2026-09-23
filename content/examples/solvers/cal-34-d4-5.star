WEEK = ["Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday", "Sunday"]

MONTHS = ["January", "February", "March", "April", "May", "June",
          "July", "August", "September", "October", "November", "December"]

def solve(options):
    start = (2024, 10, 26)
    if WEEK[weekday(start[0], start[1], start[2])] != "Saturday":
        fail("the trip does not start on a Saturday")
    end = add_days(start, 9 - 1)
    return match(options, "%s, %d %s" % (WEEK[weekday(end[0], end[1], end[2])], end[2], MONTHS[end[1] - 1]))
