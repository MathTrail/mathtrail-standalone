def verdict(answers):
    # Every way of being knights and liars that fits what was said gives one
    # answer; when they give different ones, the words do not settle it.
    if len(answers) == 0:
        fail("no mix of knights and liars can say all of this")
    return list(answers)[0] if len(answers) == 1 else "It is impossible to tell"

def solve(options):
    names = {
        (True, True): "Both are knights",
        (False, False): "Both are liars",
        (True, False): "Ben is a knight, Kim is a liar",
        (False, True): "Ben is a liar, Kim is a knight",
    }
    answers = set()
    for ann, ben, kim, said_liar in product([True, False], repeat=4):
        # Ann said either 'I am a liar' or 'I am a knight'.
        ann_says = not ann if said_liar else ann
        ben_says = said_liar  # 'Ann said that she is a liar.'
        kim_says = not ben  # 'Ben is lying.'
        if ann_says == ann and ben_says == ben and kim_says == kim:
            answers.add(names[(ben, kim)])
    return match(options, verdict(answers))
