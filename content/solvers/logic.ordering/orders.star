# For "who is where" from clues: try every order, keep the orders every clue
# allows, and read the answer off them. Orders that fit but give different
# answers mean the clues do not settle the question.
# From reference task ord-34-d2-2.

PEOPLE = ["Dan", "Eva", "Fay", "Gil", "Hal"]  # everyone, as the options write them

def fits(place):
    # Place 1 is the winner. Hal won; Dan finished after Eva but before Fay;
    # Gil finished right after Fay.
    return (place["Hal"] == 1 and place["Eva"] < place["Dan"] and place["Dan"] < place["Fay"] and
            place["Gil"] == place["Fay"] + 1)

def solve(options):
    answers = set()
    for order in permutations(PEOPLE):
        place = {name: i + 1 for i, name in enumerate(order)}
        if fits(place):
            answers.add(order[2])  # who finished third
    if len(answers) != 1:
        fail("the clues fit %d different answers" % len(answers))
    return match(options, list(answers)[0])
