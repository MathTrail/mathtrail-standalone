def solve(options):
    pattern = ["red", "red", "blue"]
    return match(options, pattern[(20 - 1) % 3])
