ROWS = (2, 4)

def moves(a, b):
    # Every move from rows of a and b counters, named as the options name it,
    # with the counters it leaves. A row is named by what it held at the start.
    found = [("%d from the row of %d" % (take, ROWS[0]), (a - take, b)) for take in range(1, a + 1)]
    found += [("%d from the row of %d" % (take, ROWS[1]), (a, b - take)) for take in range(1, b + 1)]
    found += [("%d from both rows" % take, (a - take, b - take)) for take in range(1, min(a, b) + 1)]
    return found

def solve(options):
    wins = {}
    for a in range(ROWS[0] + 1):
        for b in range(ROWS[1] + 1):
            wins[(a, b)] = any([not wins[left] for _, left in moves(a, b)])
    first = [name for name, left in moves(ROWS[0], ROWS[1]) if not wins[left]]
    if len(first) != 1:
        fail("%d first moves win, and the question asks for the one" % len(first))
    return match(options, first[0])
