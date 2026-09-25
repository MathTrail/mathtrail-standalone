ANSWERS = [1, 2, 3, 3]  # what the four islanders answered, in turn
CANNOT_TELL = "It is impossible to tell"

def verdict(answers):
    if len(answers) == 0:
        fail("no mix of knights and liars can give these answers")
    return list(answers)[0] if len(answers) == 1 else CANNOT_TELL

def solve(options):
    answers = set()
    for kinds in product([True, False], repeat=len(ANSWERS)):  # True is a knight
        knights = len([kind for kind in kinds if kind])
        # A knight gives the true number and a liar any other; at least one is a knight.
        honest = all([(ANSWERS[i] == knights) == kinds[i] for i in range(len(ANSWERS))])
        if honest and knights >= 1:
            answers.add(knights)
    return match(options, verdict(answers))
