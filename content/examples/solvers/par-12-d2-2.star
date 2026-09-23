def solve(options):
    line = ["boy", "girl"] * 5
    return match(options, len([child for child in line if child == "girl"]))
