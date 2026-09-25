MONTHS = ["January", "February", "March", "April", "May", "June", "July", "August",
          "September", "October", "November", "December"]
START = (2028, 1, 1)
DAYS = 100

def solve(options):
    if not is_leap(START[0]):
        fail("%d is not a leap year" % START[0])
    year, month, day = add_days(START, DAYS)
    return match(options, "%d %s %d" % (day, MONTHS[month - 1], year))
