BLOCK = "1234567"
TIMES = 9

def solve(options):
    # Cross out round by round, keeping the digits in even places of what is
    # left, until one round crosses out everything.
    digits = list((BLOCK * TIMES).elems())
    last = None
    while digits:
        kept = [digits[i] for i in range(len(digits)) if (i + 1) % 2 == 0]
        if not kept:
            last = digits[-1]
        digits = kept
    return match(options, int(last))
