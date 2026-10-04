# Two islanders each say something: try every way of being knights and liars,
# keep the ways in which every knight's words are true and every liar's are
# false, and read who is who off them.
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
    fitting = []
    for anya, borya in product([True, False], repeat=2):  # True is a knight
        anya_says = not borya  # 'Borya is a liar.'
        borya_says = anya and borya  # 'Anya and I are both knights.'
        if anya_says == anya and borya_says == borya:
            fitting.append((anya, borya))
    # The steps begin by supposing Borya a knight, which has to end in a contradiction.
    if [way for way in fitting if way[1]]:
        fail("Borya can be a knight")
    return match(options, verdict(set(["Anya %s, Borya %s" % (kind(a), kind(b)) for a, b in fitting])))
