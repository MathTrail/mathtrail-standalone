def solve(options):
    # Every club of up to 100 pupils who only sing, who do both and who only
    # play the drum, kept when 12 sing, one in 3 singers plays the drum and
    # one in 5 drummers sings.
    drummers = set()
    for both in range(101):
        for only_sing in range(101):
            if only_sing + both != 12 or 3 * both != only_sing + both:
                continue
            for only_drum in range(101):
                if 5 * both == both + only_drum:
                    drummers.add(both + only_drum)
    if len(drummers) != 1:
        fail("the club gives %d answers" % len(drummers))
    return match(options, list(drummers)[0])
