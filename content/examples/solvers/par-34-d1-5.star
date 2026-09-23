def solve(options):
    pattern = ["green", "yellow", "red", "yellow"]
    return match(options, pattern[(30 - 1) % 4])
