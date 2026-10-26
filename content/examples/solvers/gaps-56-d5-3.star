PEOPLE = 9
RADIUS = 40  # cm, an open umbrella round each person
APART = 60  # cm between neighbours once the umbrellas are closed

def length(apart, reach):
    # Each person stands as close to the one in front as allowed, and the
    # queue runs from the front of what the first holds out to the back of
    # what the last does.
    places = [0]
    for i in range(PEOPLE - 1):
        places.append(places[-1] + apart)
    return (places[-1] + reach) - (places[0] - reach)

def metres(cm):
    return "%d m" % (cm // 100) if cm % 100 == 0 else "%d.%d m" % (cm // 100, cm % 100 // 10)

def solve(options):
    shorter = length(2 * RADIUS, RADIUS) - length(APART, 0)
    if shorter % 10 != 0:
        fail("the answer is not a whole number of tenths of a metre")
    return match(options, metres(shorter))
