GAIN = 2  # minutes the first clock gains every hour
LOSS = 3  # minutes the second clock loses every hour
DIAL = 12 * 60  # clocks with hands show the same time 12 hours apart

def counted(number, unit):
    # "1 day", "2 days": a count as the options write it.
    return "%d %s" % (number, unit if number == 1 else unit + "s")

def solve(options):
    # The search below looks only at whole hours, so the clocks must first agree on one.
    if DIAL % (GAIN + LOSS) != 0:
        fail("the clocks first agree part of the way through an hour")
    # Go on hour by hour until the two clocks' hands stand the same way again.
    hours = 1
    while ((60 + GAIN) * hours) % DIAL != ((60 - LOSS) * hours) % DIAL:
        hours += 1
    if hours % 24 != 0:
        fail("the clocks agree at a time that is not a whole number of days")
    return match(options, counted(hours // 24, "day"))
