# How many sweets must be taken from the jar to be sure of an orange one: try
# every handful the jar allows, note the sizes of the handfuls with no orange
# sweet yet, and the answer is the smallest size none of them has.
# Written from the template draws of pigeonhole.basic.

STOCK = [8, 2]  # how many of each there are: lemon, orange

def not_yet(counts):
    return counts[1] < 1  # no orange sweet yet

def solve(options):
    failed = set()
    for counts in product(*[range(n + 1) for n in STOCK]):
        if not_yet(counts):
            failed.add(sum(counts))
    for size in range(sum(STOCK) + 1):
        if size not in failed:
            return match(options, size)
    fail("no number of sweets is ever enough")
