# The fewest steps to get exactly 1 litre with an empty 3-litre jug, an empty
# 5-litre jug and a tap: a breadth-first search over what every jug holds,
# nearest first, so the first time the goal comes up is the fewest steps.
# Written from the template pouring of algorithms.weighing_pouring.

CAPACITIES = (3, 5)  # litres each jug holds
START = (0, 0)  # litres in each jug at first
FILL = True  # whether a jug may be filled from a tap
EMPTY = True  # whether a jug may be poured out
IMPOSSIBLE = "It is impossible"  # as an option would write it

def reached(state):
    return 1 in state  # exactly 1 litre in one of the jugs

def following(state):
    # Every state one step away.
    after = []
    for i in range(len(CAPACITIES)):
        if FILL:
            after.append(state[:i] + (CAPACITIES[i],) + state[i + 1:])
        if EMPTY:
            after.append(state[:i] + (0,) + state[i + 1:])
        for j in range(len(CAPACITIES)):
            if i != j:
                amount = min([state[i], CAPACITIES[j] - state[j]])
                poured = list(state)
                poured[i] -= amount
                poured[j] += amount
                after.append(tuple(poured))
    return after

def solve(options):
    steps = {START: 0}
    queue = [START]
    head = 0
    while head < len(queue):
        state = queue[head]
        head += 1
        if reached(state):
            return match(options, steps[state])
        for after in following(state):
            if after not in steps:
                steps[after] = steps[state] + 1
                queue.append(after)
    return match(options, IMPOSSIBLE)
