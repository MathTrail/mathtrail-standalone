FIRST = ("8:00", "14:00")  # leaves A, gets to B
SECOND = ("9:00", "12:00")  # leaves B, gets to A

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
    # A road as many steps long as the two trips' minutes multiplied, so that
    # each bus covers a whole number of steps a minute.
    first_trip = at(FIRST[1]) - at(FIRST[0])
    second_trip = at(SECOND[1]) - at(SECOND[0])
    road = first_trip * second_trip
    for now in range(at(FIRST[0]), at(FIRST[1]) + 1):
        first = (now - at(FIRST[0])) * (road // first_trip)
        second = road - max(0, now - at(SECOND[0])) * (road // second_trip)
        if first == second:
            return match(options, show(now))
        if first > second:
            fail("the buses meet between two whole minutes")
    fail("the buses do not meet")
