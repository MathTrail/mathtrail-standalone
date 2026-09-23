def solve(options):
    books = ["the maths book", "the history book", "the music book", "the English book", "the art book"]
    found = set()
    for order in permutations(books):
        place = {name: i + 1 for i, name in enumerate(order)}
        if (place["the art book"] == 5 and
                abs(place["the English book"] - place["the art book"]) == 1 and
                place["the history book"] == place["the maths book"] + 1 and
                abs(place["the music book"] - place["the maths book"]) != 1):
            found.add(order[2])
    if len(found) != 1:
        fail("the clues fit %d different answers" % len(found))
    return match(options, list(found)[0])
