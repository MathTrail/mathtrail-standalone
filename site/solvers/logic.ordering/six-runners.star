# Six runners finish a race with no ties, and five clues tell how: try every
# order, keep the orders every clue allows, and read off who won.
# Written from the template orders of logic.ordering.

RUNNERS = ["Ada", "Bob", "Cid", "Dot", "Eli", "Fay"]  # as the answer names them

def fits(place):
    # Place 1 is the winner. Bob finished right after Ada; Cid finished after
    # Dot but before Ada; Fay finished right before Eli; Eli was not last;
    # Dot was not first.
    return (place["Bob"] == place["Ada"] + 1 and
            place["Dot"] < place["Cid"] and place["Cid"] < place["Ada"] and
            place["Fay"] == place["Eli"] - 1 and place["Eli"] != 6 and
            place["Dot"] != 1)

def solve(options):
    answers = set()
    for order in permutations(RUNNERS):
        place = {name: i + 1 for i, name in enumerate(order)}
        if fits(place):
            answers.add(order[0])  # who won
    if len(answers) != 1:
        fail("the clues fit %d different answers" % len(answers))
    return match(options, list(answers)[0])
