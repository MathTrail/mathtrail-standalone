def verdict(answers):
    # Every way of being knights and liars that fits what was said gives one
    # answer; when they give different ones, the words do not settle it.
    if len(answers) == 0:
        fail("no mix of knights and liars can say all of this")
    return list(answers)[0] if len(answers) == 1 else "It is impossible to tell"

def solve(options):
    size = 4
    answers = set()
    for kinds in product([True, False], repeat=size):
        fits = True
        for seat in range(size):
            left = kinds[(seat - 1) % size]
            says = not left  # 'My neighbour on the left is a liar.'
            if says != kinds[seat]:
                fits = False
        if fits:
            answers.add(len([kind for kind in kinds if kind]))
    return match(options, verdict(answers))
