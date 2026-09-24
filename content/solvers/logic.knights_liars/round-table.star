# For islanders round a table who speak about their neighbours: try every
# table of knights and liars, and keep the tables at which each one's words
# about the neighbours on the left and on the right are true exactly when the
# speaker is a knight.
# From reference task kl-34-d4-3.

SEATS = 5  # islanders at the table
CANNOT_TELL = "It is impossible to tell"  # as the option writes it, if one says so

def says(left, right):
    return not left and not right  # 'Both my neighbours are liars.'

def fits(kinds):
    for seat in range(SEATS):
        left, right = kinds[(seat - 1) % SEATS], kinds[(seat + 1) % SEATS]
        if says(left, right) != kinds[seat]:
            return False
    return True

def verdict(answers):
    # Every table that fits gives one answer; when they give different ones,
    # the words do not settle it, and when none fits, nobody could say this.
    if len(answers) == 0:
        fail("no table of knights and liars can say all of this")
    return list(answers)[0] if len(answers) == 1 else CANNOT_TELL

def solve(options):
    answers = set()
    for kinds in product([True, False], repeat=SEATS):  # True is a knight
        if fits(kinds):
            answers.add(len([kind for kind in kinds if kind]))  # how many knights
    return match(options, verdict(answers))
