def solve(options):
    price = 200
    # Each change is a percentage of the price at that moment, in hundredths.
    raised = [new for new in range(0, 1001) if new * 100 == price * (100 + 10)]
    if len(raised) != 1:
        fail("a rise of 10% on %d is not a whole price" % price)
    lowered = [new for new in range(0, raised[0] + 1) if new * 100 == raised[0] * (100 - 10)]
    if len(lowered) != 1:
        fail("a fall of 10% on %d is not a whole price" % raised[0])
    return match(options, lowered[0])
