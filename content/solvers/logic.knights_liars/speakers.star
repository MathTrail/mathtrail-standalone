# For islanders who are named and each say something: try every way of being
# knights and liars, keep the ways in which every knight's words are true and
# every liar's are false, and read the answer off them.
# From reference task kl-34-d3-5.

CANNOT_TELL = "It is impossible to tell"  # as the option writes it, if one says so

def verdict(answers):
    # Every way that fits gives one answer; when they give different ones, the
    # words do not settle it, and when none fits, nobody could say all of it.
    if len(answers) == 0:
        fail("no mix of knights and liars can say all of this")
    return list(answers)[0] if len(answers) == 1 else CANNOT_TELL

def solve(options):
    answers = set()
    for ann, ben, kim in product([True, False], repeat=3):  # True is a knight
        ann_says = not ben  # 'Ben is a liar.'
        ben_says = not kim  # 'Kim is a liar.'
        kim_says = not ann and not ben  # 'Ann and Ben are both liars.'
        if ann_says == ann and ben_says == ben and kim_says == kim:
            answers.add(len([kind for kind in [ann, ben, kim] if kind]))  # how many knights
    return match(options, verdict(answers))
