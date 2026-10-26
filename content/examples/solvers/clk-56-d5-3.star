STARTS = "13:00"
COUNTED = "14:00"
BOXES = 75

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
    # Boxes counted in sixths: in a minute each of the three packs a third of a
    # box, 2 sixths, and Zoe half a box, 3 sixths.
    found = []
    for joins in range(at(STARTS), at(COUNTED) + 1):
        sixths = 0
        for minute in range(at(STARTS), at(COUNTED)):
            sixths += 3 * 2
            if minute >= joins:
                sixths += 3
        if sixths == BOXES * 6:
            found.append(joins)
    if len(found) != 1:
        fail("%d times fit" % len(found))
    return match(options, show(found[0]))
