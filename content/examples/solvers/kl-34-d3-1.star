def verdict(answers):
    # Every way of being knights and liars that fits what was said gives one
    # answer; when they give different ones, the words do not settle it.
    if len(answers) == 0:
        fail("no mix of knights and liars can say all of this")
    return list(answers)[0] if len(answers) == 1 else "It is impossible to tell"

def solve(options):
    answers = set()
    for kinds in product([True, False], repeat=3):
        knights = len([kind for kind in kinds if kind])
        says = knights == 1  # each of them: 'Exactly one of us three is a knight.'
        # A knight's words are true and a liar's are false.
        if all([says == kind for kind in kinds]):
            answers.add(knights)
    return match(options, verdict(answers))
