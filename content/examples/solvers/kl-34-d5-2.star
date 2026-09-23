def verdict(answers):
    # Every way of being knights and liars that fits what was said gives one
    # answer; when they give different ones, the words do not settle it.
    if len(answers) == 0:
        fail("no mix of knights and liars can say all of this")
    return list(answers)[0] if len(answers) == 1 else "It is impossible to tell"

def solve(options):
    answers = set()
    for kinds in product([True, False], repeat=12):
        # Everyone after the first says: 'The person just before me is a liar.'
        if all([(not kinds[place - 1]) == kinds[place] for place in range(1, 12)]):
            answers.add(len([kind for kind in kinds if kind]))
    return match(options, verdict(answers))
