SET_RIGHT = "8:00"  # home time
GAIN = 5  # minutes the watch gains in each real hour
SHOWS = "21:00"  # the watch when she lands, later on the same day
AHEAD = 3 * 60  # the city's clocks show this many minutes more than at home

def at(clock_time):
    parts = clock_time.split(":")
    return int(parts[0]) * 60 + int(parts[1])

def show(total):
    # A time as the options write it: hours, then minutes padded by hand,
    # because % takes no width here.
    total = total % (24 * 60)
    minutes = total % 60
    return "%d:%s" % (total // 60, str(minutes) if minutes >= 10 else "0" + str(minutes))

def solve(options):
    moved = at(SHOWS) - at(SET_RIGHT)  # minutes the watch counted
    if moved <= 0:
        fail("the watch passed midnight, and the task does not say how many times")
    # The real minutes in which a watch going 60 + GAIN minutes an hour counts that many;
    # a watch that gains counts more minutes than really pass, so the search stops there.
    real = [minutes for minutes in range(1, moved + 1) if minutes * (60 + GAIN) == moved * 60]
    if len(real) != 1:
        fail("%d real times fit" % len(real))
    return match(options, show(at(SET_RIGHT) + real[0] + AHEAD))
