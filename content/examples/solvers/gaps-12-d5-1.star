def cuts_for(pieces):
    count = 1
    cuts = 0
    while count < pieces:
        cuts += 1
        count += 1
    return cuts

def solve(options):
    return match(options, 3 * cuts_for(4) * 2)
