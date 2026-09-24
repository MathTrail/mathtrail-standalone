CANNOT_TELL = "It is impossible to tell"

def verdict(answers):
    if len(answers) == 0:
        fail("no mix of knights and liars can say all of this")
    return list(answers)[0] if len(answers) == 1 else CANNOT_TELL

def solve(options):
    answers = set()
    for ivy, jay, kit in product([True, False], repeat=3):  # True is a knight
        ivy_says = jay != kit  # 'Exactly one of Jay and Kit is a knight.'
        jay_says = ivy  # 'Ivy is a knight.'
        kit_says = ivy  # 'Ivy is a knight.'
        if ivy_says == ivy and jay_says == jay and kit_says == kit:
            answers.add(len([kind for kind in [ivy, jay, kit] if kind]))  # how many knights
    return match(options, verdict(answers))
