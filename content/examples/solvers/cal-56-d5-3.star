DAYS = 366  # in the planet's year
LENGTHS = [29, 30, 31]  # at least one month of each

def solve(options):
    # Every count of months of each length, at least one of each, that makes
    # the year; each found gives a number of months.
    months = set()
    for short in range(1, DAYS // 29 + 1):
        for long in range(1, DAYS // 31 + 1):
            rest = DAYS - 29 * short - 31 * long
            if rest >= 30 and rest % 30 == 0:
                months.add(short + rest // 30 + long)
    if not months:
        fail("no months make %d days" % DAYS)
    if len(months) != 1:
        return match(options, "It cannot be told without the number of months of each length")
    return match(options, list(months)[0])
