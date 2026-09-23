def solve(options):
    pattern = ["red", "red", "white"]
    return match(options, pattern[(50 - 1) % 3])
