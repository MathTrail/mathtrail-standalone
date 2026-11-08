# A die with 2 on top rolls straight ahead over its edge twice, its opposite
# faces adding up to 7: keep its four faces round the way it rolls, and turn
# them one step a roll.
# Written from the template fold-net of geometry.cubes_paper.

TOP = 2  # the number on top at the start
ROLLS = 2  # the rolls straight ahead

def solve(options):
    # The faces round the way it rolls, top first, then front, bottom and back:
    # only the top and the bottom are known, and front and back are any pair
    # the other numbers allow.
    answers = []
    for front in [1, 2, 3, 4, 5, 6]:
        if front == TOP or front == 7 - TOP:
            continue
        ring = [TOP, front, 7 - TOP, 7 - front]
        for roll in range(ROLLS):
            ring = [ring[3], ring[0], ring[1], ring[2]]
        if ring[0] not in answers:
            answers.append(ring[0])
    if len(answers) != 1:
        return match(options, "It cannot be told")
    return match(options, answers[0])
