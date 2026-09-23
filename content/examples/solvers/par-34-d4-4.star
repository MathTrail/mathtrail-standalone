def solve(options):
    holders = []
    child = 1
    while len(holders) <= 12:
        holders.append(child)
        child = (child - 1 + 4) % 12 + 1
        if child == 1:
            return match(options, len(holders))
    fail("the ball never comes back to child 1")
