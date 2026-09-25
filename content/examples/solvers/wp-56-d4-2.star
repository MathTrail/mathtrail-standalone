CAPACITIES = (12, 7, 5)  # litres: the churn and the two jugs
START = (12, 0, 0)  # the churn full, the jugs empty
GOAL = 4  # litres wanted in one of them

def following(state):
    # Every state one step away: pour one container into another until the
    # first is empty or the second is full.
    after = []
    for i in range(len(CAPACITIES)):
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
        if GOAL in state:
            return match(options, steps[state])
        for after in following(state):
            if after not in steps:
                steps[after] = steps[state] + 1
                queue.append(after)
    fail("no container ever holds %d litres" % GOAL)
