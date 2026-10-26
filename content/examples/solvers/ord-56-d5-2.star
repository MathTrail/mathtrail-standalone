COLOURS = ["red", "blue", "green", "yellow", "white"]
PETS = ["cat", "dog", "fish", "bird", "horse"]

def solve(options):
    # Every way of painting the five houses and of housing the five pets.
    answers = set()
    for painted in permutations(COLOURS):
        house = {colour: i + 1 for i, colour in enumerate(painted)}
        if house["blue"] != 3 or house["green"] != house["white"] - 1:
            continue
        for kept in permutations(PETS):
            home = {pet: i + 1 for i, pet in enumerate(kept)}
            if home["dog"] != house["red"] or abs(home["cat"] - home["dog"]) != 1:
                continue
            if home["fish"] != 5 or abs(home["bird"] - house["yellow"]) != 1:
                continue
            answers.add("The " + kept[house["green"] - 1])
    if len(answers) != 1:
        fail("the clues leave %d pets for the green house" % len(answers))
    return match(options, list(answers)[0])
