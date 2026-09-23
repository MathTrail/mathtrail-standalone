def verdict(answers):
    # Every way of being knights and liars that fits what was said gives one
    # answer; when they give different ones, the words do not settle it.
    if len(answers) == 0:
        fail("no mix of knights and liars can say all of this")
    return list(answers)[0] if len(answers) == 1 else "It is impossible to tell"

def solve(options):
    answers = set()
    for ann, ben, kim in product([True, False], repeat=3):
        liars = len([kind for kind in [ann, ben, kim] if not kind])
        ann_says = liars == 1  # 'Exactly one of us three is a liar.'
        ben_says = liars == 2  # 'Exactly two of us three are liars.'
        kim_says = liars == 3  # 'All three of us are liars.'
        if ann_says == ann and ben_says == ben and kim_says == kim:
            answers.add(3 - liars)
    return match(options, verdict(answers))
