FACES = 6  # the numbers 1 to 6
WANTED = 3  # times one number must come up

def solve(options):
    # Every count of rolls that is not yet enough, with each number come up
    # fewer than WANTED times, built up one number at a time; the answer is
    # the first count not among them.
    failed = set([0])
    for _ in range(FACES):
        failed = set([rolls + count for rolls in failed for count in range(WANTED)])
    rolls = 0
    while rolls in failed:
        rolls += 1
    return match(options, rolls)
