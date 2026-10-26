YEAR = 2026  # any year will do: the months asked about are January to June
FIRST, LAST = 1, 6

def solve(options):
    # Every date of the months, written day first and read month first: the
    # day becomes the month and the month the day. The other reading is a
    # real date when its month, the day, is at most 12, and when its day, the
    # month's number, is in that month; and it is another date when the day
    # is not the month's own number.
    misread = 0
    for month in range(FIRST, LAST + 1):
        for day in range(1, days_in_month(YEAR, month) + 1):
            if day <= 12 and month <= days_in_month(YEAR, day) and day != month:
                misread += 1
    return match(options, misread)
