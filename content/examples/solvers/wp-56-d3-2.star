COINS = 4
REAL = 4  # the weight of a real coin
FAKE = [1, 2, 3]  # weights a fake may have: lighter, and the two fakes need not weigh the same

def on_pans(weighing):
    return len(weighing[0]) + len(weighing[1])

def weighings():
    # Every weighing: each coin on the left pan, on the right pan or aside,
    # fewest coins on the pans first. Swapping the pans tells nothing new, so
    # the pan with the lowest coin is always the left one.
    found = []
    for places in product([0, 1, 2], repeat=COINS):
        left = [coin for coin in range(COINS) if places[coin] == 1]
        right = [coin for coin in range(COINS) if places[coin] == 2]
        if left and right and left[0] < right[0]:
            found.append((left, right))
    return sorted(found, key=on_pans)

def results(worlds, weighing):
    # The worlds split by what the balance shows: left pan lighter, balanced, or heavier.
    left, right = weighing
    parts = {}
    for world in worlds:
        fakes, weights = world
        difference = sum([weights[coin] for coin in left]) - sum([weights[coin] for coin in right])
        side = 0 if difference == 0 else (1 if difference > 0 else -1)
        if side not in parts:
            parts[side] = []
        parts[side].append(world)
    return parts.values()

def settled(worlds):
    return len(set([fakes for fakes, _ in worlds])) == 1

def within_one(worlds, every):
    return settled(worlds) or any([all([settled(part) for part in results(worlds, weighing)]) for weighing in every])

def within_two(worlds, every):
    # A function may not call itself here, so each number of weighings has one of its own.
    if settled(worlds):
        return True
    for weighing in every:
        if all([within_one(part, every) for part in results(worlds, weighing)]):
            return True
    return False

def solve(options):
    worlds = []
    for fakes in combinations(range(COINS), 2):
        for pair in product(FAKE, repeat=2):
            weights = [REAL] * COINS
            weights[fakes[0]], weights[fakes[1]] = pair[0], pair[1]
            worlds.append((fakes, weights))
    every = weighings()
    if settled(worlds):
        return match(options, 0)
    if within_one(worlds, every):
        return match(options, 1)
    if within_two(worlds, every):
        return match(options, 2)
    fail("two weighings are not enough")
