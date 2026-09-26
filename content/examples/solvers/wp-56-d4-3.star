MACHINES = 10
GOOD = 10  # grams a bolt from a good machine weighs
FAULTY = 11  # grams a bolt from the faulty machine weighs

def tells_apart(counts):
    # One weighing of counts[machine] bolts from each machine finds the faulty
    # one when no two machines, being the faulty one, would give the same weight.
    weights = set()
    for faulty in range(MACHINES):
        weights.add(sum([count * (FAULTY if machine == faulty else GOOD) for machine, count in enumerate(counts)]))
    return len(weights) == MACHINES

def solve(options):
    # Two machines that give the same number of bolts weigh the same whichever
    # of them is faulty, so only different numbers can work: try every set of
    # them up to a little past one each, and keep the smallest total.
    totals = [sum(counts) for counts in combinations(range(MACHINES + 2), MACHINES) if tells_apart(counts)]
    if not totals:
        fail("no set of bolts finds the faulty machine")
    return match(options, min(totals))
