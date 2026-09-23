def solve(options):
    found = set()
    for order in permutations(["Grandma", "Mum", "Kate", "Dad"]):
        place = {name: i + 1 for i, name in enumerate(order)}
        if (place["Grandma"] < place["Mum"] and place["Mum"] < place["Kate"] and
                place["Dad"] < place["Mum"] and place["Grandma"] < place["Dad"]):
            found.add(place["Kate"] - 1)
    if len(found) != 1:
        fail("the clues fit %d different answers" % len(found))
    return match(options, list(found)[0])
