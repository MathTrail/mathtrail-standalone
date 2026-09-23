def solve(options):
    # The quotient is exact, so it is a division of whole numbers rather than a fraction.
    top = prod([2, 4, 6, 8, 10])
    bottom = prod([1, 2, 3, 4, 5])
    if top % bottom != 0:
        fail("the quotient is not a whole number")
    return match(options, top // bottom)
