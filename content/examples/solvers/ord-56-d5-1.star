def solve(options):
    # Every order the six could have finished in.
    fourths = set()
    for order in permutations(["Ann", "Ben", "Cat", "Dan", "Eve", "Fay"]):
        at = {name: i + 1 for i, name in enumerate(order)}
        if at["Cat"] < at["Dan"] and at["Dan"] < at["Eve"] and at["Ann"] == at["Dan"] + 1 and at["Ben"] == at["Eve"] + 1 and abs(at["Eve"] - at["Ann"]) != 1:
            fourths.add(order[3])
    if len(fourths) != 1:
        fail("the clues leave %d runners fourth" % len(fourths))
    return match(options, list(fourths)[0])
