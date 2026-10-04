# Three islanders each say something about the others: try every way of being
# knights and liars, keep the ways in which every knight's words are true and
# every liar's are false, and read who is who off them.
# Written from the template speakers of logic.knights_liars.

CANNOT_TELL = "It is impossible to tell"

def verdict(answers):
    # Every way that fits gives one answer; when they give different ones, the
    # words do not settle it, and when none fits, nobody could say all of it.
    if len(answers) == 0:
        fail("no mix of knights and liars can say all of this")
    return list(answers)[0] if len(answers) == 1 else CANNOT_TELL

def kind(knight):
    return "knight" if knight else "liar"

def solve(options):
    answers = set()
    for a, b, c in product([True, False], repeat=3):  # True is a knight
        a_says = not b  # 'B is a liar.'
        b_says = c  # 'C is a knight.'
        c_says = not a and not b  # 'A and B are both liars.'
        if a_says == a and b_says == b and c_says == c:
            answers.add("A %s, B %s, C %s" % (kind(a), kind(b), kind(c)))
    return match(options, verdict(answers))
