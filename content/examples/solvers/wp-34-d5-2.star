def pour_steps(capacities, start, reached, tap):
    # A breadth-first search over what is in each container, nearest first, so
    # the first time the goal comes up is the fewest steps.
    steps = {start: 0}
    queue = [start]
    head = 0
    while head < len(queue):
        state = queue[head]
        head += 1
        if reached(state):
            return steps[state]
        following = []
        for i in range(len(capacities)):
            if tap:
                following.append(state[:i] + (capacities[i],) + state[i + 1:])
                following.append(state[:i] + (0,) + state[i + 1:])
            for j in range(len(capacities)):
                if i != j:
                    amount = min([state[i], capacities[j] - state[j]])
                    poured = list(state)
                    poured[i] -= amount
                    poured[j] += amount
                    following.append(tuple(poured))
        for after in following:
            if after not in steps:
                steps[after] = steps[state] + 1
                queue.append(after)
    return None

def reached(state):
    return state[0] == 4 and state[1] == 4  # 4 litres in the bucket and 4 in the 5-litre jug

def solve(options):
    steps = pour_steps((8, 5, 3), (8, 0, 0), reached, False)
    return match(options, "It is impossible" if steps == None else steps)
