def cuts_for(pieces):
    count = 1
    cuts = 0
    while count < pieces:
        cuts += 1
        count += 1
    return cuts

def solve(options):
    cuts = cuts_for(6)
    minutes = 0
    for cut in range(1, cuts + 1):
        minutes += 5
        if cut < cuts:
            minutes += 2
    return match(options, minutes)
