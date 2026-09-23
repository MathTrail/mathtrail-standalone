def solve(options):
    boxes = 7
    wanted = 4
    items = 1
    while items < 1000:
        for most in range(items + 1):
            if boxes * most >= items:
                if most >= wanted:
                    return match(options, items)
                break
        items += 1
    fail("no number of items forces a box that full")
