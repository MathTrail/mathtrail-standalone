# Islanders round a table each say who their neighbour on the right is: try
# every table of knights and liars, keep the tables at which each one's words
# are true exactly when the speaker is a knight, and count the knights. The
# example's note says that an odd table cannot be seated at all, so a table of
# nine is searched as well, and has to fit nobody.
# Written from the template round-table of logic.knights_liars.

SEATS = 10  # islanders at the table
ODD = 9  # the table the note says no islanders can sit at
CANNOT_TELL = "It is impossible to tell"

def says(right):
    return not right  # 'My neighbour on the right is a liar.'

def fits(kinds):
    seats = len(kinds)
    for seat in range(seats):
        if says(kinds[(seat + 1) % seats]) != kinds[seat]:
            return False
    return True

def tables(seats):
    return [kinds for kinds in product([True, False], repeat=seats) if fits(kinds)]

def verdict(answers):
    # Every table that fits gives one answer; when they give different ones,
    # the words do not settle it, and when none fits, nobody could say this.
    if len(answers) == 0:
        fail("no table of knights and liars can say all of this")
    return list(answers)[0] if len(answers) == 1 else CANNOT_TELL

def solve(options):
    if len(tables(ODD)) > 0:
        fail("a table of %d fits, and the note says none does" % ODD)
    answers = set([len([kind for kind in kinds if kind]) for kinds in tables(SEATS)])
    return match(options, verdict(answers))
