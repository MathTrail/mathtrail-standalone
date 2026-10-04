# Two islanders meet on a road and one of them speaks: try every way of being
# knights and liars, keep the ways in which the speaker's words are true
# exactly when the speaker is a knight, and read who is who off them.
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
    for a, b in product([True, False], repeat=2):  # True is a knight
        a_says = not a and b  # 'I am a liar, and B is a knight.'
        if a_says == a:
            answers.add("A %s, B %s" % (kind(a), kind(b)))
    return match(options, verdict(answers))
