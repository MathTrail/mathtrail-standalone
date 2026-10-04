# A game on a strip of cells, where the players in turn put a counter on an
# empty cell that touches no counter, and whoever cannot loses: work out, from
# the full strip down to the empty one, which positions win for the player
# about to move, then find the first moves that leave the other a losing one.
# Written from the template one-pile of games.strategy.

CELLS = 9
MIDDLE = CELLS // 2  # counted from 0

def can_put(taken, cell):
    # A cell takes a counter when neither it nor a neighbour holds one.
    for near in [cell - 1, cell, cell + 1]:
        if near >= 0 and near < CELLS and (taken >> near) & 1:
            return False
    return True

def mirrored(taken):
    flipped = 0
    for cell in range(CELLS):
        if (taken >> cell) & 1:
            flipped = flipped | (1 << (CELLS - 1 - cell))
    return flipped

def solve(options):
    # wins[taken] says whether the player about to move wins. A counter only
    # adds a bit, so every position a move leads to is worked out first.
    wins = [False] * (1 << CELLS)
    for taken in range((1 << CELLS) - 1, -1, -1):
        wins[taken] = any([not wins[taken | (1 << cell)] for cell in range(CELLS) if can_put(taken, cell)])
    first = [cell + 1 for cell in range(CELLS) if not wins[1 << cell]]
    if len(first) != 1:
        fail("%d first moves win, and the question asks for the one" % len(first))
    # The steps say that once the middle is taken, a mirror reply is always free.
    for taken in range(1 << CELLS):
        if (taken >> MIDDLE) & 1 and mirrored(taken) == taken:
            for cell in range(CELLS):
                if can_put(taken, cell) and not can_put(taken | (1 << cell), CELLS - 1 - cell):
                    fail("after a counter on cell %d its mirror is not free" % (cell + 1))
    return match(options, first[0])
