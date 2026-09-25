CANNOT_TELL = "It is impossible to tell"

def verdict(answers):
    if len(answers) == 0:
        fail("no order of places fits what was said")
    return list(answers)[0] if len(answers) == 1 else CANNOT_TELL

def solve(options):
    found = set()
    for order in permutations(["Tess", "Ugo", "Vera", "Walt"]):
        place = {name: i + 1 for i, name in enumerate(order)}
        # Each of the three said one true thing and one false thing.
        tess = (place["Tess"] == 1) != (place["Ugo"] == 2)
        ugo = (place["Ugo"] == 1) != (place["Tess"] == 3)
        vera = (place["Vera"] == 2) != (place["Walt"] == 1)
        if tess and ugo and vera:
            found.add(order[0])
    return match(options, verdict(found))
