def joined(group):
    # Walk from one slab to its neighbours by whole sides; the group stays
    # joined when the walk reaches every slab of it.
    seen = [group[0]]
    head = 0
    while head < len(seen):
        row, column = seen[head]
        head += 1
        for step_row, step_column in [(-1, 0), (1, 0), (0, -1), (0, 1)]:
            neighbour = (row + step_row, column + step_column)
            if neighbour in group and neighbour not in seen:
                seen.append(neighbour)
    return len(seen) == len(group)

def solve(options):
    slabs = [(row, column) for row in range(3) for column in range(3) if (row, column) != (1, 1)]
    splits = 0
    for group in combinations(slabs, 4):
        # The first slab always goes to the first gardener, so each split is
        # counted once rather than once for each gardener.
        if slabs[0] not in group:
            continue
        other = [slab for slab in slabs if slab not in group]
        if joined(list(group)) and joined(other):
            splits += 1
    return match(options, splits)
