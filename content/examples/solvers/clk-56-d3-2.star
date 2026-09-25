DAYS = ["Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday", "Sunday"]
LEAVES = ("Monday", "8:00")  # on the clocks in P
ARRIVES = ("Tuesday", "20:00")  # on the clocks in Q
AHEAD = 5 * 60  # the clocks in Q show this many minutes more than in P

def at(clock_time):
    parts = clock_time.split(":")
    return int(parts[0]) * 60 + int(parts[1])

def moment(day, clock_time):
    # Minutes from the start of the week.
    return DAYS.index(day) * 24 * 60 + at(clock_time)

def solve(options):
    # Both moments on the clocks in P.
    voyage = moment(ARRIVES[0], ARRIVES[1]) - AHEAD - moment(LEAVES[0], LEAVES[1])
    if voyage % 60 != 0:
        fail("the voyage is not a whole number of hours")
    return match(options, voyage // 60)
