def solve(options):
    # Every order the five could stand in.
    thirds = set()
    for order in permutations(["Ali", "Bea", "Cal", "Dina", "Eli"]):
        at = {name: i + 1 for i, name in enumerate(order)}
        if at["Bea"] < at["Dina"] and at["Dina"] < at["Ali"] and at["Cal"] == at["Ali"] + 1 and at["Eli"] != 5 and abs(at["Eli"] - at["Bea"]) != 1:
            thirds.add(order[2])
    if len(thirds) != 1:
        fail("the clues leave %d children third" % len(thirds))
    return match(options, list(thirds)[0])
