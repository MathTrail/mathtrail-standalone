CANNOT_TELL = "It is impossible to tell"

def verdict(answers):
    # Orders that fit but put different girls fifth mean the clues do not settle it.
    if len(answers) == 0:
        fail("no order of finishing fits the clues")
    return list(answers)[0] if len(answers) == 1 else CANNOT_TELL

def solve(options):
    found = set()
    for order in permutations(["Lara", "Maya", "Nell", "Olive", "Paula", "Rosa"]):
        place = {name: i + 1 for i, name in enumerate(order)}
        if (place["Olive"] > place["Maya"] and place["Rosa"] < place["Maya"] and
                place["Nell"] < place["Maya"] and place["Paula"] > place["Olive"] and
                place["Paula"] != place["Olive"] + 1):
            found.add(order[4])  # who finished fifth
    return match(options, verdict(found))
