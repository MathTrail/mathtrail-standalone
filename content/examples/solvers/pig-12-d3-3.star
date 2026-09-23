def solve(options):
    items = 9
    boxes = 4
    smallest = items
    for counts in product(range(items + 1), repeat=boxes):
        if sum(counts) == items:
            smallest = min([smallest, max(counts)])
    return match(options, smallest)
