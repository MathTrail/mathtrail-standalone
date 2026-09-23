def solve(options):
    trees = [5 * i for i in range(21)]
    return match(options, trees[-1] - trees[0])
