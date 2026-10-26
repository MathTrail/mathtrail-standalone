def solve(options):
    # Every whole number of minutes Max could be later than Lily: at their one
    # speed his walk, a fifth of the way out, a fifth back home and then the
    # whole way, is seven fifths of what Lily walks in her 50 minutes.
    way = 50
    fifths = 1 + 1 + 5
    later = [m for m in range(0, 201) if (way + m) * 5 == way * fifths]
    if len(later) != 1:
        fail("%d delays fit" % len(later))
    return match(options, later[0])
