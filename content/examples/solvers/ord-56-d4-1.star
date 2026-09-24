CANNOT_TELL = "It is impossible to tell"

def verdict(answers):
    if len(answers) == 0:
        fail("no rack of bikes fits the clues")
    return list(answers)[0] if len(answers) == 1 else CANNOT_TELL

def solve(options):
    found = set()
    for order in permutations(["Hana", "Ivo", "Jon", "Kai", "Lea", "Max", "Ned"]):
        slot = {name: i + 1 for i, name in enumerate(order)}  # slot 1 is on the left
        if (slot["Hana"] == 3 and abs(slot["Ivo"] - slot["Jon"]) - 1 == 3 and
                slot["Kai"] == slot["Jon"] + 1 and slot["Hana"] < slot["Lea"] and
                slot["Lea"] < slot["Ivo"]):
            found.add(slot["Ivo"])
    return match(options, verdict(found))
