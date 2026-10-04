# A daisy of 12 petals in a ring, and two players who take turns tearing off
# one petal or two petals next to each other; whoever tears off the last one
# wins. The first move opens the ring into a row of 11 or 10 petals, and a row
# loses petals the same way, which can split it into two shorter rows. Work
# out which positions win for the player about to move, fewest petals first, a
# position being the lengths of its rows, longest first. The steps say two
# equal rows of 5 or of 4 lose for the player about to move, so that is
# checked too.
# Written from the template two-piles of games.strategy.

PETALS = 12

def position(rows):
    # The lengths of the rows left, without empty ones, longest first.
    return tuple(sorted([row for row in rows if row > 0], reverse=True))

def moves(state):
    # Every position one move away: one or two petals torn from some row,
    # which leaves the petals on either side of them as two rows.
    after = []
    for at in range(len(state)):
        others = list(state[:at]) + list(state[at + 1:])
        row = state[at]
        for torn in [1, 2]:
            for left in range(0, row - torn + 1):
                after.append(position(others + [left, row - torn - left]))
    return after

def solve(options):
    # positions[n] is every position of n petals in rows, longest row first.
    positions = {0: [()]}
    for n in range(1, PETALS):
        positions[n] = [(first,) + rest for first in range(1, n + 1) for rest in positions[n - first]
                        if len(rest) == 0 or rest[0] <= first]
    # wins[position] says whether the player about to move can force a win.
    # With no petal left, the other player tore off the last one.
    wins = {(): False}
    for n in range(1, PETALS):
        for state in positions[n]:
            wins[state] = any([not wins[after] for after in moves(state)])
    for equal in [(5, 5), (4, 4)]:
        if wins[equal]:
            fail("two rows of %d win for the player about to move" % equal[0])
    first_wins = not wins[(PETALS - 1,)] or not wins[(PETALS - 2,)]
    return match(options, "The first player" if first_wins else "The second player")
