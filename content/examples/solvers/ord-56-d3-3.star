CANNOT_TELL = "It is impossible to tell"

def verdict(answers):
    # Rows that fit but end with different friends mean the clues do not settle it.
    if len(answers) == 0:
        fail("no row of friends fits the clues")
    return list(answers)[0] if len(answers) == 1 else CANNOT_TELL

def solve(options):
    found = set()
    for order in permutations(["Alice", "Ben", "Dev", "Ella", "Kira"]):
        place = {name: i + 1 for i, name in enumerate(order)}  # place 1 is on the left
        if (place["Alice"] == 1 and abs(place["Kira"] - place["Dev"]) != 1 and
                abs(place["Dev"] - place["Ella"]) != 1 and place["Kira"] < place["Dev"]):
            found.add(order[-1])
    return match(options, verdict(found))
