PAIR = {
    (True, True): "Both are knights",
    (False, False): "Both are liars",
    (True, False): "Ann is a knight, Ben is a liar",
    (False, True): "Ann is a liar, Ben is a knight",
}

def verdict(answers):
    # Every way of being knights and liars that fits what was said gives one
    # answer; when they give different ones, the words do not settle it.
    if len(answers) == 0:
        fail("no mix of knights and liars can say all of this")
    return list(answers)[0] if len(answers) == 1 else "It is impossible to tell"

def solve(options):
    answers = set()
    for ann, ben in product([True, False], repeat=2):
        ann_says = not ann or not ben  # 'At least one of us two is a liar.'
        # A knight's words are true and a liar's are false; Ben says nothing.
        if ann_says == ann:
            answers.add(PAIR[(ann, ben)])
    return match(options, verdict(answers))
