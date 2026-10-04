# A chess tournament of 5 rounds of 50 minutes, with a 15-minute break between
# every two rounds, starts at 10:00: count the breaks as pairs of neighbouring
# rounds, add the play and the breaks to the start, and turn the end back into
# a time with show. The steps count 310 minutes in all, so that is checked
# too.
# Written from the template minutes of time.clocks.

STARTS = "10:00"
ROUNDS = 5
ROUND = 50  # minutes of play in one round
BREAK = 15  # minutes between two rounds

def at(clock_time):
    parts = clock_time.split(":")
    return int(parts[0]) * 60 + int(parts[1])

def show(total):
    # A time as the answer writes it: hours, then minutes padded by hand,
    # because % takes no width here.
    total = total % (24 * 60)
    minutes = total % 60
    return "%d:%s" % (total // 60, str(minutes) if minutes >= 10 else "0" + str(minutes))

def solve(options):
    breaks = len([(round, round + 1) for round in range(1, ROUNDS)])
    lasts = ROUNDS * ROUND + breaks * BREAK
    if lasts != 310:
        fail("the tournament lasts %d minutes, and the steps say 310" % lasts)
    return match(options, show(at(STARTS) + lasts))
