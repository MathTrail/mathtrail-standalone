# For a game on two piles: work out which pairs of piles win for the player
# about to move, smallest first, then find the first moves that hand the
# other player a losing pair. Every kind of move the rules allow goes into
# moves(), and a move only ever makes the piles smaller.
# From reference task game-56-d4-2.

PILES = (2, 4)  # what the piles hold at the start; the options tell them apart by it
FROM_ONE = "%d from the row of %d"  # how an option names a move from one pile: how many, from which
FROM_BOTH = "%d from both rows"  # how an option names taking the same from both; "" when the rules forbid it

def moves(a, b):
    # Every move from piles of a and b, named as the options name it, with the
    # piles it leaves.
    found = [(FROM_ONE % (take, PILES[0]), (a - take, b)) for take in range(1, a + 1)]
    found += [(FROM_ONE % (take, PILES[1]), (a, b - take)) for take in range(1, b + 1)]
    if FROM_BOTH:
        found += [(FROM_BOTH % take, (a - take, b - take)) for take in range(1, min(a, b) + 1)]
    return found

def solve(options):
    # wins[(a, b)] says whether the player about to move can force a win; the
    # piles a move leaves are smaller, so they are known by the time they are
    # needed. With nothing left to take, the player to move has lost.
    wins = {}
    for a in range(PILES[0] + 1):
        for b in range(PILES[1] + 1):
            wins[(a, b)] = any([not wins[left] for _, left in moves(a, b)])
    first = [name for name, left in moves(PILES[0], PILES[1]) if not wins[left]]
    if len(first) != 1:
        fail("%d first moves win, and the question asks for the one" % len(first))
    return match(options, first[0])
