def solve(options):
    children = ["Ann", "Ben", "Kim"]
    return match(options, children[(10 - 1) % 3])
