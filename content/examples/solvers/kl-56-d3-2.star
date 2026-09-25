PEOPLE = 7  # islanders in the line, the first at the front

def solve(options):
    answers = set()
    for kinds in product([True, False], repeat=PEOPLE):  # True is a knight
        fits = True
        for place in range(PEOPLE):
            ahead = kinds[:place]
            knights = len([kind for kind in ahead if kind])
            says = len(ahead) - knights > knights  # 'In front of me there are more liars than knights.'
            if says != kinds[place]:
                fits = False
        if fits:
            answers.add(len([kind for kind in kinds if kind]))  # how many knights
    if len(answers) != 1:
        fail("the words fit %d different answers" % len(answers))
    return match(options, list(answers)[0])
