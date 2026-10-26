SETS_OFF = "10:40"
TURNS = "11:05"

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
    # Every river speed and every speed of the boat through the water that
    # lets it go up the river, in steps a minute; up the river the boat goes
    # its speed less the current's, and down it, its speed and the current's.
    met = set()
    for current in range(4):
        for speed in range(current + 1, current + 6):
            ball, boat, now = 0, 0, at(SETS_OFF)
            reached = None
            while now < at(TURNS) + 24 * 60:
                now += 1
                ball += current
                boat += current - speed if now <= at(TURNS) else current + speed
                if now > at(TURNS) and boat >= ball:
                    if boat != ball:
                        fail("the boat reaches the ball between two whole minutes")
                    reached = now
                    break
            if reached == None:
                fail("the boat does not reach the ball within a day")
            met.add(reached)
    if len(met) != 1:
        return match(options, "It depends on how fast the river flows")
    return match(options, "At " + show(list(met)[0]))
