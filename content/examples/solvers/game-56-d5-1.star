def solve(options):
    # The children between the players on each side, fewer first: standing
    # next to the other player at one's own turn wins, and otherwise a player
    # sends home a neighbour from one side or the other.
    wins = {}
    for total in range(0, 9):
        for one_side in range(0, total + 1):
            other_side = total - one_side
            if one_side == 0 or other_side == 0:
                wins[(one_side, other_side)] = True
            else:
                wins[(one_side, other_side)] = not wins[(one_side - 1, other_side)] or not wins[(one_side, other_side - 1)]
    if wins[(3, 5)]:
        return match(options, "Ann, since she goes first")
    return match(options, "Ben")
