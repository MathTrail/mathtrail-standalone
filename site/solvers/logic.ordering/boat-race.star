# Three boats race, and two clues tell how they finished: try every order,
# keep the orders both clues allow, and read off which boat won.
# Written from the template orders of logic.ordering.

BOATS = ["red", "blue", "green"]  # as the answer names them

def fits(place):
    # Place 1 is the winner. The red boat finished after the blue one; the
    # green boat finished before the blue one.
    return place["red"] > place["blue"] and place["green"] < place["blue"]

def solve(options):
    answers = set()
    for order in permutations(BOATS):
        place = {boat: i + 1 for i, boat in enumerate(order)}
        if fits(place):
            answers.add(order[0])  # who won
    if len(answers) != 1:
        fail("the clues fit %d different answers" % len(answers))
    return match(options, list(answers)[0])
