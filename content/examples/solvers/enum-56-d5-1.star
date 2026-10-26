BEADS = 8
BLUE = 2

def turned(necklace):
    # Every way the loop can be turned round, and turned over as well.
    out = []
    for shift in range(BEADS):
        moved = necklace[shift:] + necklace[:shift]
        out.append(moved)
        out.append(moved[::-1])
    return out

def solve(options):
    seen = {}
    different = 0
    for places in combinations(range(BEADS), BLUE):
        necklace = tuple(["blue" if i in places else "red" for i in range(BEADS)])
        if necklace in seen:
            continue
        different += 1
        for same in turned(necklace):
            seen[same] = True
    return match(options, different)
