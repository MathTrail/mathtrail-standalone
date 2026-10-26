OPENS = "7:00"
MIX = 20  # minutes
RISE = 120
BAKE = 40
COOL = 15
HEAT = 30  # the oven, which can heat while the dough rises

def at(clock_time):
    parts = clock_time.split(":")
    return int(parts[0]) * 60 + int(parts[1])

def show(total):
    # A time as the options write it: hours, then minutes padded by hand,
    # because % takes no width here.
    total = total % (24 * 60)
    minutes = total % 60
    return "%d:%s" % (total // 60, str(minutes) if minutes >= 10 else "0" + str(minutes))

def ready(start, oven):
    # The dough is mixed and rises from the start, the oven heats for 30
    # minutes from the minute it is lit, and baking begins once the dough has
    # risen and the oven is hot.
    bakes = max(start + MIX + RISE, oven + HEAT)
    return bakes + BAKE + COOL

def solve(options):
    # The starts from 7:00 back, the first that works being the latest. The
    # baker lights the oven on starting, since lighting it later could only
    # hold the baking back.
    for start in range(at(OPENS), -1, -1):
        if ready(start, start) <= at(OPENS):
            return match(options, show(start))
    fail("no start is early enough")
