def seatings(size, says):
    # Every way the table can be knights and liars in which each one's words,
    # said about the neighbour on the left and the one on the right, are true
    # exactly when the speaker is a knight.
    found = []
    for kinds in product([True, False], repeat=size):
        if all([says(kinds[(seat - 1) % size], kinds[(seat + 1) % size]) == kinds[seat] for seat in range(size)]):
            found.append(kinds)
    return found

def both_liars(left, right):
    return not left and not right  # 'Both my neighbours are liars.'

def solve(options):
    found = seatings(10, both_liars)
    if len(found) == 0:
        fail("no table of ten can say this")
    return match(options, max([len([kind for kind in kinds if kind]) for kinds in found]))
