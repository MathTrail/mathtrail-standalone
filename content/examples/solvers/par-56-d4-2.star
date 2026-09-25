LAMPS = 8
RUN = 3  # lamps next to each other that one move switches
NEVER = "It can never be done"  # as the option writes it

def fewest_moves():
    # A breadth-first search over which lamps are on, nearest first, so the
    # first time all of them are on is the fewest moves.
    start = tuple([0] * LAMPS)
    goal = tuple([1] * LAMPS)
    moves = {start: 0}
    queue = [start]
    head = 0
    while head < len(queue):
        state = queue[head]
        head += 1
        if state == goal:
            return moves[state]
        for first in range(LAMPS - RUN + 1):
            switched = tuple([1 - lamp if first <= place and place < first + RUN else lamp for place, lamp in enumerate(state)])
            if switched not in moves:
                moves[switched] = moves[state] + 1
                queue.append(switched)
    return None

def solve(options):
    moves = fewest_moves()
    return match(options, NEVER if moves == None else moves)
