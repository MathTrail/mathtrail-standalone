def solve(options):
    # Every whole distance Ann could have walked when Tom has walked 12 km, the
    # two always walking 2 km for every 3 km of Tom's.
    km = [a for a in range(0, 101) if a * 3 == 12 * 2]
    if len(km) != 1:
        fail("%d distances fit" % len(km))
    return match(options, "%d km" % km[0])
