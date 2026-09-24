# For a game on one pile, where the players move in turn and each move takes
# one of a few allowed amounts: work out, from the empty pile upwards, which
# piles win for the player about to move, then find the first moves that hand
# the other player a losing pile.
# From reference task game-56-d4-1.

PILE = 15  # what the pile holds at the start
TAKES = [1, 3, 4]  # the amounts a move may take
LAST_LOSES = False  # True when whoever takes the last one loses
CANNOT_WIN = "She cannot win for certain"  # as the option writes it, if one says so

def the_move(first):
    # The one first move that wins, when the question asks which: CANNOT_WIN
    # when none does, and a stop when several do.
    if len(first) == 0:
        return CANNOT_WIN
    if len(first) > 1:
        fail("%d first moves win, and the question asks for the one" % len(first))
    return first[0]

def solve(options):
    # wins[n] says whether the player about to move with n left can force a
    # win. With none left the other player took the last one, so the player
    # to move has lost, unless taking the last one loses.
    wins = [LAST_LOSES]
    for n in range(1, PILE + 1):
        wins.append(any([not wins[n - take] for take in TAKES if take <= n]))
    first = [take for take in TAKES if take <= PILE and not wins[PILE - take]]
    return match(options, the_move(first))  # len(first) when the question asks how many
