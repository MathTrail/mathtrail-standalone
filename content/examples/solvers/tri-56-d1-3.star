DIGITS = [1, 2, 3, 4, 5, 6]
TOTAL = 120

def solve(options):
    # The digits added in turn until the total is reached; it must be hit
    # exactly, not passed.
    total = 0
    step = 0
    digit = None
    while total < TOTAL:
        digit = DIGITS[step % len(DIGITS)]
        total += digit
        step += 1
    if total != TOTAL:
        fail("the total is passed, not reached")
    return match(options, digit)
