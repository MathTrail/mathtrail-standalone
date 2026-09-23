def solve(options):
    names = ["Ira", "Kolya", "Lida", "Misha"]
    found = set()
    for values in permutations([128, 131, 134, 137]):
        height = dict(zip(names, values))
        if (height["Ira"] > height["Kolya"] and height["Kolya"] > height["Lida"] and
                height["Misha"] == 131 and height["Misha"] > height["Lida"]):
            found.add(height["Ira"])
    if len(found) != 1:
        fail("the clues fit %d different answers" % len(found))
    return match(options, list(found)[0])
