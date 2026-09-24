JARS = (3, 5)

def moves(a, b):
    # Every move from jars of a and b coins, named as the options name it,
    # with the coins it leaves. A jar is named by what it held at the start.
    found = [("%d from the jar of %d" % (take, JARS[0]), (a - take, b)) for take in range(1, a + 1)]
    found += [("%d from the jar of %d" % (take, JARS[1]), (a, b - take)) for take in range(1, b + 1)]
    return found

def solve(options):
    # wins[(a, b)] says whether the player about to move can force a win;
    # every move leaves fewer coins, so those positions are known already.
    wins = {}
    for a in range(JARS[0] + 1):
        for b in range(JARS[1] + 1):
            wins[(a, b)] = any([not wins[left] for _, left in moves(a, b)])
    first = [name for name, left in moves(JARS[0], JARS[1]) if not wins[left]]
    if len(first) != 1:
        fail("%d first moves win, and the question asks for the one" % len(first))
    return match(options, first[0])
