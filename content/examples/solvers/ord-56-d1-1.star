def solve(options):
    # Every order the five could have finished in.
    places = set()
    for order in permutations(["Kim", "Leo", "Max", "Nina", "Olga"]):
        at = {name: i + 1 for i, name in enumerate(order)}
        if at["Leo"] == at["Nina"] + 1 and at["Olga"] < at["Kim"] and at["Kim"] < at["Max"] and at["Max"] != 5:
            places.add(at["Max"])
    if len(places) != 1:
        fail("the clues leave %d places for Max" % len(places))
    return match(options, list(places)[0])
