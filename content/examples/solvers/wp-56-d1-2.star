JUGS = (4, 6)  # litres
CANDIDATES = [2, 1, 3, 5, 10]  # the amounts the options name, in litres

def written(litres):
    return "1 litre" if litres == 1 else "%d litres" % litres

def solve(options):
    # Every state the steps can reach from two empty jugs: fill one, empty one,
    # or pour one into the other until it is empty or the other is full.
    seen = set([(0, 0)])
    queue = [(0, 0)]
    head = 0
    while head < len(queue):
        a, b = queue[head]
        head += 1
        into_b = min(a, JUGS[1] - b)
        into_a = min(b, JUGS[0] - a)
        for state in [(JUGS[0], b), (a, JUGS[1]), (0, b), (a, 0), (a - into_b, b + into_b), (a + into_a, b - into_a)]:
            if state not in seen:
                seen.add(state)
                queue.append(state)
    amounts = set([a for a, _ in seen] + [b for _, b in seen])
    fitting = [litres for litres in CANDIDATES if litres in amounts]
    if len(fitting) != 1:
        fail("%d of the amounts can be made" % len(fitting))
    return match(options, written(fitting[0]))
