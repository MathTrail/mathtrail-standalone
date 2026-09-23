def solve(options):
    items = 4
    boxes = 3
    smallest = items
    for counts in product(range(items + 1), repeat=boxes):
        if sum(counts) == items:
            smallest = min([smallest, max(counts)])
    return match(options, smallest)
