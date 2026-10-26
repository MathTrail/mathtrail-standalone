DEPTH = 12  # metres
UP = 5  # each day
DOWN = 3  # each night

def solve(options):
    height = 0
    day = 0
    while day < 100:
        day += 1
        height += UP
        if height >= DEPTH:
            return match(options, "On day %d" % day)
        height -= DOWN
    fail("the frog is still in the well after 100 days")
