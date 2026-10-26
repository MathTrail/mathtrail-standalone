WEIGHTS = [10, 20, 30, 40, 50, 60, 70, 80]  # grams
CANDIDATES = [  # what each option says can be left at the end
    ("10 g", [10]),
    ("80 g", [80]),
    ("70 g", [70]),
    ("Any of them but 80 g", [10, 20, 30, 40, 50, 60, 70]),
    ("Any of them", WEIGHTS),
]
TAKEN, LEFT, RIGHT = 0, 1, 2  # where a weight is

def solve(options):
    # Every start, each weight on one pan or the other with the pans unequal,
    # and from each state every weight a child can take: one that leaves the
    # other pan the heavier. A state carries the grams on each pan with it.
    queue = []
    for places in product([LEFT, RIGHT], repeat = len(WEIGHTS)):
        left = sum([WEIGHTS[i] for i in range(len(WEIGHTS)) if places[i] == LEFT])
        right = sum(WEIGHTS) - left
        if left != right:
            queue.append((places, left, right))
    seen = set([places for places, _, _ in queue])
    can_be_left = set()
    head = 0
    while head < len(queue):
        places, left, right = queue[head]
        head += 1
        on = [i for i in range(len(WEIGHTS)) if places[i] != TAKEN]
        if len(on) == 1:
            can_be_left.add(WEIGHTS[on[0]])
            continue
        for i in on:
            new_left = left - WEIGHTS[i] if places[i] == LEFT else left
            new_right = right - WEIGHTS[i] if places[i] == RIGHT else right
            if new_left == new_right or (new_left > new_right) == (left > right):
                continue
            after = places[:i] + (TAKEN,) + places[i + 1:]
            if after not in seen:
                seen.add(after)
                queue.append((after, new_left, new_right))
    fitting = [text for text, weights in CANDIDATES if sorted(can_be_left) == sorted(weights)]
    if len(fitting) != 1:
        fail("%d of the options say what can be left" % len(fitting))
    return match(options, fitting[0])
