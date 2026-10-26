YEAR = 2024

MONTHS = ["January", "February", "March", "April", "May", "June", "July",
          "August", "September", "October", "November", "December"]

def digit_sum(number):
    return sum([int(digit) for digit in str(number).elems()])

def solve(options):
    # Every date of the year in order; the last whose day and month have
    # digits adding up to as much as the year's.
    last = None
    for month in range(1, 13):
        for day in range(1, days_in_month(YEAR, month) + 1):
            if digit_sum(day) + digit_sum(month) == digit_sum(YEAR):
                last = (day, month)
    if last == None:
        fail("no date of %d fits" % YEAR)
    return match(options, "%d %s" % (last[0], MONTHS[last[1] - 1]))
