TAKES_OFF = "10:00"  # by the clocks where it takes off
FLIGHT = 3 * 60  # minutes
BEHIND = 4 * 60  # the clocks where it lands show this many minutes less

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
    return match(options, show(at(TAKES_OFF) + FLIGHT - BEHIND))
