def solve(options):
    pencils = ["the white pencil", "the yellow pencil", "the red pencil", "the green pencil", "the blue pencil"]
    found = set()
    for order in permutations(pencils):
        place = {name: i + 1 for i, name in enumerate(order)}
        if (place["the yellow pencil"] < place["the red pencil"] and
                place["the red pencil"] < place["the green pencil"] and
                place["the green pencil"] < place["the blue pencil"] and
                place["the white pencil"] < place["the yellow pencil"]):
            found.add(order[2])
    if len(found) != 1:
        fail("the clues fit %d different answers" % len(found))
    return match(options, list(found)[0])
