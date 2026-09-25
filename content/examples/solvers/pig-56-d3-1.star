STOCK = [1, 2, 9]  # coins of 50, 20 and 10 cents
WANTED = 3  # coins of one value

def solve(options):
    failed = set()
    for counts in product(*[range(n + 1) for n in STOCK]):
        if max(counts) < WANTED:  # no value has three yet
            failed.add(sum(counts))
    for size in range(sum(STOCK) + 1):
        if size not in failed:
            return match(options, size)
    fail("no number of coins is ever enough")
