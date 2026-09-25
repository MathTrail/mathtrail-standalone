AGO = 8  # years ago
AHEAD = 4  # years from now
TIMES = 3  # her age AHEAD years from now is this many times her age AGO years ago

def fits(age):
    # Her age now, and from it her ages then and later.
    then, later = age - AGO, age + AHEAD
    return then >= 0 and later == TIMES * then

def solve(options):
    fitting = [age for age in range(120) if fits(age)]
    if len(fitting) != 1:
        fail("%d ages fit the clue" % len(fitting))
    return match(options, fitting[0])
