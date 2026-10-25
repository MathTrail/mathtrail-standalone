NIGHTS = 4

def solve(options):
    # She is away on the day she goes, and each night away ends in a day of its
    # own, the last of them the day she comes back.
    days = ["the day she goes"]
    for night in range(1, NIGHTS + 1):
        days.append("the day after night %d" % night)
    return match(options, len(days))
