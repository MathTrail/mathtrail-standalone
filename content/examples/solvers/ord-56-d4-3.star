def solve(options):
    found = set()
    for order in permutations(["Abby", "Bob", "Cora", "Dean", "Emma", "Finn"]):
        act = {name: i + 1 for i, name in enumerate(order)}
        if (act["Emma"] in (1, 3) and abs(act["Finn"] - act["Emma"]) == 1 and act["Bob"] > act["Cora"] and
                act["Dean"] in (2, 6) and abs(act["Abby"] - act["Bob"]) - 1 == 1):
            found.add(order[2])  # who performs third
    if len(found) != 1:
        fail("the clues fit %d different answers" % len(found))
    return match(options, list(found)[0])
