COINS = 9
FAKES = 5  # next to each other round the circle, lighter than the real coins
CANDIDATES = [([1], [5]), ([1], [2]), ([1], [4]), ([1, 2], [3, 4]), ([1, 3], [5, 7])]  # the weighings the options name

def weight(coin, fakes):
    return 1 if coin in fakes else 2

def shows_two(left, right):
    # Every place the run of fakes can start, sorted by what the weighing then
    # shows; for each result, the coins that are fakes in all the runs it
    # leaves must be two at least.
    runs_by_result = {}
    for start in range(COINS):
        fakes = [(start + i) % COINS + 1 for i in range(FAKES)]
        difference = sum([weight(coin, fakes) for coin in left]) - sum([weight(coin, fakes) for coin in right])
        result = 0 if difference == 0 else (1 if difference > 0 else -1)
        runs_by_result.setdefault(result, []).append(fakes)
    for runs in runs_by_result.values():
        sure = [coin for coin in range(1, COINS + 1) if all([coin in fakes for fakes in runs])]
        if len(sure) < 2:
            return False
    return True

def written(weighing):
    left, right = weighing
    if len(left) == 1:
        return "Coin %d against coin %d" % (left[0], right[0])
    return "Coins %d and %d against coins %d and %d" % (left[0], left[1], right[0], right[1])

def solve(options):
    fitting = [weighing for weighing in CANDIDATES if shows_two(weighing[0], weighing[1])]
    if len(fitting) != 1:
        fail("%d of the weighings are sure to show two fakes" % len(fitting))
    return match(options, written(fitting[0]))
