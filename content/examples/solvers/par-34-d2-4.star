def fewest_moves(items, per_move):
    # A breadth-first search over which items are still the wrong way up,
    # nearest first, so the first time the goal comes up is the fewest moves.
    start = tuple([1] * items)
    goal = tuple([0] * items)
    moves = {start: 0}
    queue = [start]
    head = 0
    while head < len(queue):
        state = queue[head]
        head += 1
        if state == goal:
            return moves[state]
        for chosen in combinations(range(items), per_move):
            turned = tuple([1 - side if index in chosen else side for index, side in enumerate(state)])
            if turned not in moves:
                moves[turned] = moves[state] + 1
                queue.append(turned)
    return None

def solve(options):
    moves = fewest_moves(7, 2)
    return match(options, "It can never be done" if moves == None else moves)
