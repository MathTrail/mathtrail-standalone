WANTED = 2  # pupils with the same sum

def solve(options):
    sums = set([number // 10 + number % 10 for number in range(10, 100)])  # of the two digits
    # Every size a group can have with fewer than WANTED pupils on each sum,
    # built up one sum at a time; the answer is the first size not among them.
    failed = set([0])
    for _ in sums:
        failed = set([size + count for size in failed for count in range(WANTED)])
    pupils = 0
    while pupils in failed:
        pupils += 1
    return match(options, pupils)
