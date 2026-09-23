def solve(options):
    found = set()
    for order in permutations(["Kolya", "Dima", "r1", "r2", "r3", "r4"]):
        place = {name: i + 1 for i, name in enumerate(order)}
        behind_kolya = len(order) - place["Kolya"]
        ahead_of_dima = place["Dima"] - 1
        if behind_kolya == 4 and ahead_of_dima == 3:
            found.add(abs(place["Dima"] - place["Kolya"]) - 1)
    if len(found) != 1:
        fail("the clues fit %d different answers" % len(found))
    return match(options, list(found)[0])
