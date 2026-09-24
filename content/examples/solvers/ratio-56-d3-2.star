def solve(options):
    big, small, turns = 21, 7, 6
    # Tooth for tooth: the small wheel's turns times its teeth equal the big wheel's.
    found = [t for t in range(0, 1001) if t * small == turns * big]
    if len(found) != 1:
        fail("%d numbers of turns fit" % len(found))
    return match(options, found[0])
