CAPACITIES = (4, 7)  # litres each pot holds
GOAL = 6  # litres wanted in one pot
IMPOSSIBLE = "It is impossible"  # as the option writes it

def following(state):
    # Every state one step away: fill a pot from the tap, or pour one pot into
    # the other until the first is empty or the second is full. No water is
    # ever poured away.
    after = []
    for i in range(len(CAPACITIES)):
        after.append(state[:i] + (CAPACITIES[i],) + state[i + 1:])
        j = 1 - i
        amount = min([state[i], CAPACITIES[j] - state[j]])
        poured = list(state)
        poured[i] -= amount
        poured[j] += amount
        after.append(tuple(poured))
    return after

def solve(options):
    # Breadth first from two empty pots; a search that runs out of states
    # without meeting the goal means it cannot be done.
    steps = {(0, 0): 0}
    queue = [(0, 0)]
    head = 0
    while head < len(queue):
        state = queue[head]
        head += 1
        if GOAL in state:
            return match(options, steps[state])
        for after in following(state):
            if after not in steps:
                steps[after] = steps[state] + 1
                queue.append(after)
    return match(options, IMPOSSIBLE)
