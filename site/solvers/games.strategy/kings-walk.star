# A king in the bottom left corner of a 5 by 5 board moves one square up, one
# square right, or one square diagonally up and to the right, and whoever moves
# it into the top right corner wins: work out which squares win for the player
# about to move, nearest the corner first, a square being how many rows and
# how many columns it still is from that corner. The steps say a square loses
# exactly when both of those are even, so that is checked square by square.
# Written from the template two-piles of games.strategy.

SIZE = 5
START = (SIZE - 1, SIZE - 1)  # rows and columns from the corner at the start

def moves(rows, columns):
    # Every square one move nearer the corner.
    found = []
    if rows > 0:
        found.append((rows - 1, columns))
    if columns > 0:
        found.append((rows, columns - 1))
    if rows > 0 and columns > 0:
        found.append((rows - 1, columns - 1))
    return found

def solve(options):
    # wins[square] says whether the player about to move can force a win.
    # With the king in the corner, the other player has just won.
    wins = {}
    for rows in range(SIZE):
        for columns in range(SIZE):
            wins[(rows, columns)] = any([not wins[left] for left in moves(rows, columns)])
            if wins[(rows, columns)] == (rows % 2 == 0 and columns % 2 == 0):
                fail("the square %d rows and %d columns away does not lose exactly when both are even" % (rows, columns))
    return match(options, "The first player" if wins[START] else "The second player")
