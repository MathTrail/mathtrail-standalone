# 23 sweets in a bowl, each move takes 1, 2, 3 or 4 of them, and whoever takes
# the last one wins: work out, from the empty bowl upwards, which bowls win for
# the player about to move, then find the first moves that hand the other
# player a losing bowl. The note says that after Kim takes 4, Leo can leave
# 15, a losing bowl, so that is checked too.
# Written from the template one-pile of games.strategy.

PILE = 23  # sweets in the bowl at the start
TAKES = [1, 2, 3, 4]  # the amounts a move may take

def solve(options):
    # wins[n] says whether the player about to move with n left can force a
    # win. With none left the other player took the last one.
    wins = [False]
    for n in range(1, PILE + 1):
        wins.append(any([not wins[n - take] for take in TAKES if take <= n]))
    if wins[15] or not wins[PILE - 4]:
        fail("after Kim takes 4, Leo cannot leave her a losing bowl of 15")
    first = [take for take in TAKES if take <= PILE and not wins[PILE - take]]
    if len(first) != 1:
        fail("%d first moves win, and the question asks for the one" % len(first))
    return match(options, first[0])
