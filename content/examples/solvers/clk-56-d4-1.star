THERE = ("10:00", "15:00")  # takes off by Alpha's clocks, lands by Beta's, all on one day
BACK = ("17:00", "18:00")  # takes off by Beta's clocks, lands by Alpha's

def at(clock_time):
    parts = clock_time.split(":")
    return int(parts[0]) * 60 + int(parts[1])

def counted(number, unit):
    # "1 hour", "2 hours": a count as the options write it.
    return "%d %s" % (number, unit if number == 1 else unit + "s")

def solve(options):
    off_alpha, on_beta = at(THERE[0]), at(THERE[1])
    off_beta, on_alpha = at(BACK[0]), at(BACK[1])
    found = set()
    for flight in range(1, 24 * 60):  # minutes in the air, the same both ways
        ahead = on_beta - off_alpha - flight  # Beta's clocks minus Alpha's, by the flight there
        if off_beta + flight - ahead == on_alpha:  # and the flight back agrees
            found.add(ahead)
    if len(found) != 1:
        fail("%d differences fit" % len(found))
    ahead = list(found)[0]
    if ahead == 0 or ahead % 60 != 0:
        fail("the clocks do not differ by a whole number of hours")
    return match(options, counted(abs(ahead) // 60, "hour") + (" ahead" if ahead > 0 else " behind"))
