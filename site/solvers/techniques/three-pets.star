# Three children and three pets, one each, and clues saying who has not got
# which: try every way of sharing the pets out, keep the ways every clue
# allows, and read the answer off them. Ways that fit but give different
# answers mean the clues do not settle the question.
# Written from the template orders of logic.ordering.

CHILDREN = ["Anya", "Borya", "Vera"]
PETS = ["cat", "dog", "parrot"]

def fits(pet):
    # Anya does not keep the cat; Borya is afraid of dogs; the cat does not
    # live with Borya.
    return pet["Anya"] != "cat" and pet["Borya"] != "dog" and pet["Borya"] != "cat"

def solve(options):
    answers = set()
    for shared in permutations(PETS):
        pet = {child: shared[i] for i, child in enumerate(CHILDREN)}
        if fits(pet):
            answers.add(", ".join(["%s %s" % (child, pet[child]) for child in CHILDREN]))
    if len(answers) != 1:
        fail("the clues fit %d different answers" % len(answers))
    return match(options, list(answers)[0])
