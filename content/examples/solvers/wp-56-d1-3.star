QUARTERS = 4  # an amount is counted in quarters of a kilogram
RICE = 11 * QUARTERS
WEIGHT = 1 * QUARTERS
PART = 3 * QUARTERS  # the smaller part; the rest is the 8 kg

def splits(amount):
    # What one weighing makes of rice shared between the pans until they
    # balance: two equal halves; with the weight on one pan, two parts that
    # differ by the weight; or, with the weight alone against it, the weight's
    # worth and the rest. Parts that are not whole quarters are left out.
    found = []
    if amount % 2 == 0:
        found.append((amount // 2, amount // 2))
    if amount > WEIGHT and (amount - WEIGHT) % 2 == 0:
        found.append(((amount - WEIGHT) // 2, (amount + WEIGHT) // 2))
    if amount > WEIGHT:
        found.append((WEIGHT, amount - WEIGHT))
    return found

def done(parts):
    # Parts can be poured together freely: done when some of them make the
    # smaller part.
    sums = set([0])
    for part in parts:
        sums = sums | set([s + part for s in sums])
    return PART in sums

def solve(options):
    # The parts after each number of weighings, found breadth first. Before a
    # weighing, any of the parts may be poured together.
    states = set([(RICE,)])
    for weighings in range(4):
        if any([done(parts) for parts in states]):
            return match(options, weighings)
        later = set()
        for parts in states:
            for chosen in product([False, True], repeat = len(parts)):
                poured = sum([parts[i] for i in range(len(parts)) if chosen[i]])
                if poured == 0:
                    continue
                rest = [parts[i] for i in range(len(parts)) if not chosen[i]]
                for a, b in splits(poured):
                    later.add(tuple(sorted(rest + [a, b])))
        states = later
    fail("3 kg is not reached in the weighings searched")
