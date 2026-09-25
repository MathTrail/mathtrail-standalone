MARBLES = 12  # one of them a lighter fake
ROOM = 2  # marbles that fit on one pan

def solve(options):
    # k marbles on each pan leave the fake among the k on the pan that rises,
    # or among the rest when the pans balance; the answer for every number of
    # marbles is filled in from the bottom up, because a function may not call
    # itself here, and no pan takes more than ROOM.
    best = {0: 0, 1: 0}
    for count in range(2, MARBLES + 1):
        best[count] = min([1 + max([best[k], best[count - 2 * k]]) for k in range(1, min([ROOM, count // 2]) + 1)])
    return match(options, best[MARBLES])
