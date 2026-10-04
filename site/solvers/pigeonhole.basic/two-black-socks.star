# How many socks must be taken in the dark to be sure of two black ones: try
# every handful the drawer allows, note the sizes of the handfuls with fewer
# than two black socks, and the answer is the smallest size none of them has.
# Written from the template draws of pigeonhole.basic.

STOCK = [8, 6, 2]  # how many of each there are: black, white, striped
WANTED = 2  # how many black socks the question wants to be sure of

def not_yet(counts):
    return counts[0] < WANTED  # fewer than two black socks yet

def solve(options):
    failed = set()
    for counts in product(*[range(n + 1) for n in STOCK]):
        if not_yet(counts):
            failed.add(sum(counts))
    for size in range(sum(STOCK) + 1):
        if size not in failed:
            return match(options, size)
    fail("no number of socks is ever enough")
