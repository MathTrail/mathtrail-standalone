# For the fewest moves to a position, or whether it can be reached at all: a
# breadth-first search over the positions, nearest first, so the first time
# the goal comes up is the fewest moves. A search that runs out of positions
# without meeting it means it can never be done.
# From reference task par-34-d4-2.

COINS = 5  # coins in a row, all heads up at the start
PER_MOVE = 3  # coins turned over in one move
NEVER = "It can never be done"  # as the option writes it

def fewest_moves():
    start = tuple([1] * COINS)  # 1 is heads, 0 is tails
    goal = tuple([0] * COINS)
    moves = {start: 0}
    queue = [start]
    head = 0
    while head < len(queue):
        state = queue[head]
        head += 1
        if state == goal:
            return moves[state]
        for chosen in combinations(range(COINS), PER_MOVE):
            turned = tuple([1 - side if place in chosen else side for place, side in enumerate(state)])
            if turned not in moves:
                moves[turned] = moves[state] + 1
                queue.append(turned)
    return None

def solve(options):
    moves = fewest_moves()
    return match(options, NEVER if moves == None else moves)
