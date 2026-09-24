CANNOT_TELL = "It is impossible to tell"

def verdict(answers):
    if len(answers) == 0:
        fail("no row of stalls fits the clues")
    return list(answers)[0] if len(answers) == 1 else CANNOT_TELL

def solve(options):
    found = set()
    for order in permutations(["toys", "candy", "cakes", "popcorn", "juice"]):
        number = {stall: i + 1 for i, stall in enumerate(order)}  # stall 1 is nearest the gate
        if (number["toys"] == 4 and abs(number["candy"] - number["cakes"]) == 1 and
                number["cakes"] > number["popcorn"]):
            found.add(number["popcorn"])
    return match(options, verdict(found))
