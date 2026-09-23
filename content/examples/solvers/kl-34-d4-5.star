def verdict(answers):
    # Every way of being knights and liars that fits what was said gives one
    # answer; when they give different ones, the words do not settle it.
    if len(answers) == 0:
        fail("no mix of knights and liars can say all of this")
    return list(answers)[0] if len(answers) == 1 else "It is impossible to tell"

def seatings(size, says):
    # Every way the table can be knights and liars in which each one's words,
    # said about the neighbour on the left and the one on the right, are true
    # exactly when the speaker is a knight.
    found = []
    for kinds in product([True, False], repeat=size):
        if all([says(kinds[(seat - 1) % size], kinds[(seat + 1) % size]) == kinds[seat] for seat in range(size)]):
            found.append(kinds)
    return found

def right_knight_left_liar(left, right):
    return right and not left  # 'My neighbour on the right is a knight, and my neighbour on the left is a liar.'

def solve(options):
    answers = set([len([kind for kind in kinds if kind]) for kinds in seatings(6, right_knight_left_liar)])
    return match(options, verdict(answers))
