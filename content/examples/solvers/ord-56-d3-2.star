SISTERS = ["Amy", "Beth", "Cleo", "Dora", "Gwen"]
CANNOT_TELL = "It is impossible to tell"

def verdict(answers):
    if len(answers) == 0:
        fail("no order of ages fits the clues")
    return list(answers)[0] if len(answers) == 1 else CANNOT_TELL

def solve(options):
    found = set()
    for order in permutations(SISTERS):
        place = {name: i + 1 for i, name in enumerate(order)}  # place 1 is the eldest
        younger_than_amy = [sister for sister in SISTERS if place[sister] > place["Amy"]]
        older_than_beth = [sister for sister in SISTERS if place[sister] < place["Beth"]]
        if (len(younger_than_amy) == 1 and len(older_than_beth) == 1 and
                place["Gwen"] < place["Cleo"] and place["Cleo"] < place["Amy"]):
            found.add(order[0])
    return match(options, verdict(found))
