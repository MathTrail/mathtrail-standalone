def solve(options):
    sides = ["boy", "girl"]
    return match(options, len([place for place in range(1, 16) if sides[(place - 1) % 2] == "boy"]))
