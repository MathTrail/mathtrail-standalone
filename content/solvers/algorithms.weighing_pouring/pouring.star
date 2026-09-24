# For the fewest steps of pouring: a breadth-first search over what every jug
# holds, nearest first, so the first time the goal comes up is the fewest
# steps. A search that runs out of states without meeting it means it cannot
# be done.
# From reference task wp-34-d1-1.

CAPACITIES = (3, 5)  # litres each jug holds
START = (0, 0)  # litres in each jug at first
TAP = True  # whether a jug may be filled from a tap and emptied onto the ground
IMPOSSIBLE = "It is impossible"  # as the option writes it

def reached(state):
    return 2 in state  # exactly 2 litres in one of the jugs

def following(state):
    # Every state one step away.
    after = []
    for i in range(len(CAPACITIES)):
        if TAP:
            after.append(state[:i] + (CAPACITIES[i],) + state[i + 1:])
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
