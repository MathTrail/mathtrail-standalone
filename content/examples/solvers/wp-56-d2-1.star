CAPACITIES = (3, 5, 8)  # litres each bucket holds
START = (0, 0, 0)
GOAL = 2  # litres wanted in the 8-litre bucket

def following(state):
    # Every state one step away: fill a bucket from the tap, or pour one bucket
    # into another until the first is empty or the second is full.
    after = []
    for i in range(len(CAPACITIES)):
        after.append(state[:i] + (CAPACITIES[i],) + state[i + 1:])
        for j in range(len(CAPACITIES)):
            if i != j:
                amount = min([state[i], CAPACITIES[j] - state[j]])
                poured = list(state)
                poured[i] -= amount
                poured[j] += amount
                after.append(tuple(poured))
    return after

def solve(options):
    # Breadth first, nearest states first, so the first time the goal comes up
    # is the fewest steps.
    steps = {START: 0}
    queue = [START]
    head = 0
    while head < len(queue):
        state = queue[head]
        head += 1
        if state[2] == GOAL:
            return match(options, steps[state])
        for after in following(state):
            if after not in steps:
                steps[after] = steps[state] + 1
                queue.append(after)
    fail("the 8-litre bucket never holds %d litres" % GOAL)
