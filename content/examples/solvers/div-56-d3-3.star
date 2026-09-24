def solve(options):
    number = 5555555
    # The remainder is the r from 0 to 8 that leaves number - r a multiple of 9.
    remainders = [r for r in range(0, 9) if (number - r) % 9 == 0]
    if len(remainders) != 1:
        fail("%d remainders fit" % len(remainders))
    return match(options, remainders[0])
