def verdict(answers):
    # Every way of being knights and liars that fits what was said gives one
    # answer; when they give different ones, the words do not settle it.
    if len(answers) == 0:
        fail("no mix of knights and liars can say all of this")
    return list(answers)[0] if len(answers) == 1 else "It is impossible to tell"

def solve(options):
    answers = set()
    for ann, ben, kim in product([True, False], repeat=3):
        ann_says = ben  # 'Ben is a knight.'
        ben_says = kim  # 'Kim is a knight.'
        kim_says = 2 + 2 == 5  # '2 + 2 = 5.'
        # A knight's words are true and a liar's are false.
        if ann_says == ann and ben_says == ben and kim_says == kim:
            answers.add(len([kind for kind in [ann, ben, kim] if kind]))
    return match(options, verdict(answers))
