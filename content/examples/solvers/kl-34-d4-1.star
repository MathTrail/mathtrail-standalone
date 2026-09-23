def verdict(answers):
    # Every way of being knights and liars that fits what was said gives one
    # answer; when they give different ones, the words do not settle it.
    if len(answers) == 0:
        fail("no mix of knights and liars can say all of this")
    return list(answers)[0] if len(answers) == 1 else "It is impossible to tell"

def solve(options):
    answers = set()
    for kinds in product([True, False], repeat=5):
        liars = len([kind for kind in kinds if not kind])
        # The one in place p says: 'Exactly p of us are liars.'
        if all([(liars == place + 1) == kinds[place] for place in range(5)]):
            answers.add(5 - liars)
    return match(options, verdict(answers))
