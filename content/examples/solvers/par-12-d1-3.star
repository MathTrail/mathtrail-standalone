def solve(options):
    children = 9
    while children >= 2:
        children -= 2
    return match(options, children)
