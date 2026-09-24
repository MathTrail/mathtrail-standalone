# For "how many must you take to be sure": try every handful the stock allows,
# note the sizes of the handfuls that are not yet enough, and the answer is the
# smallest size that none of them has. Taking more never undoes what is wanted,
# which is what makes the smallest such size the answer.
# From reference task pig-12-d4-1.

STOCK = [4, 5, 6]  # how many of each colour there are: red, blue, green
WANTED = 3  # how many of one colour the question wants to be sure of

def not_yet(counts):
    return max(counts) < WANTED  # no colour has three yet

def solve(options):
    failed = set()
    for counts in product(*[range(n + 1) for n in STOCK]):
        if not_yet(counts):
            failed.add(sum(counts))
    for size in range(sum(STOCK) + 1):
        if size not in failed:
            return match(options, size)
    fail("no number of draws is ever enough")
