def solve(options):
    # The face repeats every twelve hours, so the clock is right again once it
    # has gained a whole turn of it.
    days = 1
    while days < 10000:
        if (6 * days) % (12 * 60) == 0:
            return match(options, days)
        days += 1
    fail("the clock never shows the right time again")
